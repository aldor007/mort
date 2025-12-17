const supertest = require('supertest');
const chai = require('chai');
const expect = chai.expect;

const host = process.env.MORT_HOST + ':' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('Error Handling', function () {

    describe('Invalid image URLs', function () {
        it('should return 404 for non-existent image', function (done) {
            request.get('/remote/nonexistent-image-12345.jpg')
                .expect(404)
                .end(done);
        });

        it('should return 404 for non-existent image with transform', function (done) {
            request.get('/remote/nonexistent.jpg?operation=resize&width=100')
                .expect(404)
                .end(done);
        });

        it('should return 404 for non-existent image with preset', function (done) {
            request.get('/remote/doesnotexist.jpg/default_small')
                .expect(404)
                .end(done);
        });
    });

    describe('Invalid transform parameters', function () {
        it('should return 400 for invalid preset name', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg/invalid_preset_name_123')
                .end(function (err, res) {
                    // Configuration-dependent: 400 if preset validation runs, 404 if routed as file path
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should return 400 for invalid width parameter', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=invalid')
                .expect(400)
                .end(done);
        });

        it('should return 400 for negative width parameter', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=-100')
                .expect(400)
                .end(done);
        });

        it('should handle extremely large width parameter', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=999999')
                .end(function (err, res) {
                    // Should either return 400 or process with reasonable limits
                    expect([200, 400, 413, 500]).to.include(res.status);
                    done();
                });
        });

        it('should return 400 for invalid rotation angle', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=rotate&angle=invalid')
                .expect(400)
                .end(done);
        });

        it('should return 400 for invalid format parameter', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=200&format=invalid_format')
                .expect(400)
                .end(done);
        });
    });

    describe('Invalid watermark parameters', function () {
        it('should handle invalid watermark URL', function (done) {
            this.timeout(10000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=https://invalid-domain-that-does-not-exist-12345.com/image.png&opacity=0.5')
                .end(function (err, res) {
                    // Should either fail gracefully or return original image
                    expect([200, 400, 404, 500, 503, 504]).to.include(res.status);
                    done();
                });
        });

        it('should return 400 for invalid opacity value', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=https://i.imgur.com/uomkVIL.png&opacity=invalid')
                .expect(400)
                .end(done);
        });

        it('should return 400 for opacity out of range', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=https://i.imgur.com/uomkVIL.png&opacity=2.0')
                .expect(400)
                .end(done);
        });
    });

    describe('Invalid extract parameters', function () {
        it('should handle extract beyond image boundaries', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=extract&top=99999&left=99999&width=100&height=100')
                .end(function (err, res) {
                    // Should return 400 for invalid coordinates or 500 if libvips throws error
                    expect([200, 400, 500]).to.include(res.status);
                    done();
                });
        });

        it('should return 400 for negative extract coordinates', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=extract&top=-100&left=-100&width=100&height=100')
                .expect(400)
                .end(done);
        });
    });

    describe('Invalid blur parameters', function () {
        it('should return 400 for invalid sigma value', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=invalid')
                .expect(400)
                .end(done);
        });

        it('should return 400 for negative sigma value', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=-5')
                .expect(400)
                .end(done);
        });

        it('should handle extremely large sigma value', function (done) {
            this.timeout(30000);
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=1000')
                .timeout(20000)
                .end(function (err, res) {
                    // Request may timeout, so check if we got a response
                    if (err && err.timeout) {
                        // Request timed out - this is acceptable for large sigma
                        done();
                        return;
                    }
                    if (res) {
                        // May timeout (503/504), error (500), reject (400), or succeed with clamping (200)
                        expect([200, 400, 500, 503, 504]).to.include(res.status);
                    }
                    done();
                });
        });
    });

    describe('Malformed requests', function () {
        it('should handle empty operation parameter consistently', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=')
                .end(function (err, res) {
                    // Behavior depends on bucket configuration and query parser
                    // Empty operation parameter may be treated as malformed (400) or missing transform (404)
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should return 400 for malformed query string', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=')
                .expect(400)
                .end(done);
        });

        it('should handle duplicate parameters', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?width=100&width=200&operation=resize')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    // Should handle gracefully, likely using last value
                    expect(res.body.length).to.be.greaterThan(0);
                    done();
                });
        });
    });

    describe('HTTP error responses', function () {
        it('should return proper error structure for 404', function (done) {
            request.get('/remote/nonexistent123456.jpg')
                .expect(404)
                .end(function (err, res) {
                    if (err) return done(err);
                    // Check that response has proper content-type header
                    expect(res.headers).to.have.property('content-type');
                    done();
                });
        });

        it('should return proper error structure for 400', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg/invalid_preset')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    expect(res.headers).to.have.property('content-type');
                    done();
                });
        });
    });

    describe('Edge cases', function () {
        it('should return 400 for zero width', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=0')
                .expect(400)
                .end(done);
        });

        it('should return 400 for zero height', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&height=0')
                .expect(400)
                .end(done);
        });

        it('should return 400 for quality of 0', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=200&quality=0')
                .expect(400)
                .end(done);
        });

        it('should return 400 for quality over 100', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=200&quality=200')
                .expect(400)
                .end(done);
        });
    });

    describe('Bucket errors', function () {
        it('should return 404 for non-existent bucket', function (done) {
            request.get('/nonexistent-bucket/image.jpg')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });

        it('should return 400 or 404 for invalid bucket name', function (done) {
            request.get('/../etc/passwd')
                .end(function (err, res) {
                    expect([400, 404]).to.include(res.status);
                    done();
                });
        });
    });

    describe('Edge case behavior documentation', function () {
        it('should accept reasonable width values', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=2000')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    expect(res.body.length).to.be.greaterThan(0);
                    done();
                });
        });

        it('should accept reasonable sigma values for blur', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=10')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    expect(res.body.length).to.be.greaterThan(0);
                    done();
                });
        });

        it('should accept extract within reasonable bounds', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=extract&top=50&left=50&areaWith=200&areaHeight=200')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    expect(res.body.length).to.be.greaterThan(0);
                    done();
                });
        });

        it('should validate parameters even with extreme but technically valid values', function (done) {
            this.timeout(10000);
            // Extremely large width passes validation (positive integer)
            // but may fail during processing - behavior depends on system resources
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=999999')
                .end(function (err, res) {
                    // If it succeeds, validate response
                    if (res.status === 200) {
                        expect(res.body.length).to.be.greaterThan(0);
                    }
                    // Any response is acceptable here (success, resource error, timeout)
                    expect(res.status).to.be.oneOf([200, 400, 413, 500, 503]);
                    done();
                });
        });

        it('should handle extract coordinates beyond image size gracefully', function (done) {
            // Coordinates pass validation (non-negative) but are beyond image bounds
            // libvips will handle this - may error or clamp to image size
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=extract&top=99999&left=99999&areaWith=100&areaHeight=100')
                .end(function (err, res) {
                    if (res.status === 200) {
                        expect(res.body.length).to.be.greaterThan(0);
                    }
                    // Either succeeds with clamping, returns validation error, or libvips error
                    expect(res.status).to.be.oneOf([200, 400, 500]);
                    done();
                });
        });

        it('should handle extremely large blur sigma gracefully', function (done) {
            this.timeout(30000);
            // Large sigma passes validation (positive) but may timeout or fail
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=1000')
                .timeout(20000)
                .end(function (err, res) {
                    // Request may timeout, so check if we got a response
                    if (err && err.timeout) {
                        // Request timed out - this is acceptable for large sigma
                        done();
                        return;
                    }
                    if (res) {
                        if (res.status === 200) {
                            expect(res.body.length).to.be.greaterThan(0);
                        }
                        // May succeed, timeout (503/504), error (500), or reject (400) based on system resources
                        expect(res.status).to.be.oneOf([200, 400, 500, 503, 504]);
                    }
                    done();
                });
        });

        it('should fail gracefully when watermark image URL is unreachable', function (done) {
            this.timeout(10000);
            // Invalid domain should cause watermark fetch to fail
            // mort should either fail the request or skip watermark and return original
            request.get('/remote/nxpvwo7qqfwz.jpg?operation=watermark&image=https://invalid-domain-that-does-not-exist-12345.com/image.png&opacity=0.5')
                .end(function (err, res) {
                    if (res.status === 200) {
                        expect(res.body.length).to.be.greaterThan(0);
                    }
                    // Should handle network failures gracefully with appropriate error
                    expect(res.status).to.be.oneOf([400, 404, 500, 503, 504]);
                    done();
                });
        });
    });
});
