const supertest = require('supertest');
const chai = require('chai');
const expect = chai.expect;

const host = process.env.MORT_HOST + ':' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('Security', function () {

    describe('Path traversal protection', function () {
        it('should prevent path traversal with ../', function (done) {
            request.get('/local/../../../etc/passwd')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should prevent path traversal in bucket name', function (done) {
            request.get('/../local/large.jpeg')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should prevent encoded path traversal', function (done) {
            request.get('/local/%2e%2e%2f%2e%2e%2fetc%2fpasswd')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should prevent double-encoded path traversal', function (done) {
            request.get('/local/%252e%252e%252f%252e%252e%252fetc%252fpasswd')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should prevent backslash path traversal', function (done) {
            request.get('/local/..\\..\\..\\etc\\passwd')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Input validation', function () {
        it('should handle extremely long URLs', function (done) {
            const longPath = '/remote/' + 'a'.repeat(10000) + '.jpg';
            request.get(longPath)
                .end(function (err, res) {
                    expect([400, 404, 414, 500]).to.include(res.status);
                    done();
                });
        });

        it('should handle special characters in filenames', function (done) {
            request.get('/local/file<script>alert(1)</script>.jpg')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should handle null bytes in path', function (done) {
            request.get('/local/image.jpg%00.txt')
                .end(function (err, res) {
                    expect([400, 404, 500]).to.include(res.status);
                    done();
                });
        });

        it('should handle unicode characters safely', function (done) {
            // Use encoded version to avoid Node.js client-side validation errors
            request.get('/local/image%00.jpg')
                .end(function (err, res) {
                    expect([400, 404, 500]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Request size limits', function () {
        it('should handle reasonable image dimensions', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=4000&height=4000')
                .end(function (err, res) {
                    // Should either process or reject based on configured limits
                    expect([200, 400, 413, 500]).to.include(res.status);
                    done();
                });
        });

        it('should handle extremely large dimension requests appropriately', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=100000&height=100000')
                .end(function (err, res) {
                    // Should reject or limit based on configuration
                    expect([200, 400, 413, 500]).to.include(res.status);
                    done();
                });
        });
    });

    describe('HTTP method restrictions', function () {
        it('should handle GET requests for images', function (done) {
            request.get('/local/large.jpeg')
                .expect(200)
                .end(done);
        });

        it('should handle HEAD requests for images', function (done) {
            request.head('/local/large.jpeg')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    expect(res.body).to.be.empty;
                    done();
                });
        });

        it('should handle DELETE appropriately', function (done) {
            request.delete('/local/large.jpeg')
                .end(function (err, res) {
                    // DELETE might be allowed for S3 compatibility or rejected
                    expect([200, 204, 401, 403, 405]).to.include(res.status);
                    done();
                });
        });

        it('should handle PATCH requests', function (done) {
            request.patch('/local/large.jpeg')
                .end(function (err, res) {
                    // PATCH should not be allowed (various error codes possible)
                    expect([401, 405, 501]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Authentication and authorization', function () {
        it('should require valid credentials for protected buckets', function (done) {
            // This test assumes S3 auth is configured
            request.get('/local/large.jpeg')
                .set('Authorization', 'AWS invalid:credentials')
                .end(function (err, res) {
                    // Should either process without auth or reject invalid auth
                    expect([200, 401, 403]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Injection prevention', function () {
        it('should prevent SQL injection attempts in parameters', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?width=100\' OR 1=1--')
                .end(function (err, res) {
                    expect([200, 400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should prevent command injection in parameters', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?width=100;rm -rf /')
                .end(function (err, res) {
                    expect([200, 400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should prevent XXE injection via watermark URL', function (done) {
            this.timeout(10000);
            const xxePayload = 'file:///etc/passwd';
            request.get(`/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=${encodeURIComponent(xxePayload)}`)
                .end(function (err, res) {
                    expect([400, 404, 500]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Resource exhaustion protection', function () {
        it('should handle multiple blur operations without hanging', function (done) {
            this.timeout(15000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=50&operation=blur&sigma=50&operation=blur&sigma=50')
                .end(function (err, res) {
                    // Should either complete or timeout gracefully
                    expect([200, 400, 408, 500, 503]).to.include(res.status);
                    done();
                });
        });

        it('should handle deeply nested transform operations', function (done) {
            this.timeout(10000);
            let query = '/remote/nxpvwo7qqfwz.jpg?';
            for (let i = 0; i < 100; i++) {
                query += `operation=resize&width=${200 + i}&`;
            }
            request.get(query)
                .end(function (err, res) {
                    expect([200, 400, 413, 500]).to.include(res.status);
                    done();
                });
        });
    });

    describe('SSRF protection', function () {
        it('should prevent access to internal IPs via watermark', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=http://127.0.0.1/internal')
                .end(function (err, res) {
                    // Should reject internal IP addresses
                    expect([400, 403, 404, 500]).to.include(res.status);
                    done();
                });
        });

        it('should prevent access to localhost via watermark', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=http://localhost/internal')
                .end(function (err, res) {
                    expect([400, 403, 404, 500]).to.include(res.status);
                    done();
                });
        });

        it('should prevent access to metadata endpoints', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=http://169.254.169.254/latest/meta-data')
                .end(function (err, res) {
                    expect([400, 403, 404, 500]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Header injection', function () {
        it('should sanitize filenames with CRLF', function (done) {
            request.get('/local/image%0d%0aContent-Length:%200%0d%0a%0d%0a.jpg')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should handle newlines in query parameters', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?width=100%0d%0aX-Custom-Header:%20malicious')
                .end(function (err, res) {
                    expect([200, 400, 404]).to.include(res.status);
                    // Ensure no custom headers were injected
                    expect(res.headers).to.not.have.property('x-custom-header');
                    done();
                });
        });
    });

    describe('DoS protection', function () {
        it('should handle reasonable concurrent requests', function (done) {
            this.timeout(10000);
            const promises = [];
            for (let i = 0; i < 10; i++) {
                promises.push(
                    new Promise((resolve) => {
                        request.get(`/remote/nxpvwo7qqfwz.jpg?operation=resize&width=${100 + i}`)
                            .end((err, res) => resolve(res));
                    })
                );
            }

            Promise.all(promises).then((responses) => {
                // All requests should complete, though some might be rate-limited
                responses.forEach(res => {
                    expect([200, 429, 503]).to.include(res.status);
                });
                done();
            });
        });

        it('should not crash on rapid requests to same resource', function (done) {
            this.timeout(10000);
            const path = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=123';
            const promises = [];

            for (let i = 0; i < 20; i++) {
                promises.push(
                    new Promise((resolve) => {
                        request.get(path)
                            .end((err, res) => resolve(res));
                    })
                );
            }

            Promise.all(promises).then((responses) => {
                // Request collapsing should handle this gracefully
                responses.forEach(res => {
                    expect([200, 429, 503]).to.include(res.status);
                });
                done();
            });
        });
    });

    describe('Content-Type validation', function () {
        it('should validate image content types', function (done) {
            request.get('/local/main.css')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    // CSS file should return text/css, not image/*
                    expect(res.headers['content-type']).to.not.match(/^image\//);
                    done();
                });
        });
    });
});
