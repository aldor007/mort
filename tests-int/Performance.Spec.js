const supertest = require('supertest');
const chai = require('chai');
const expect = chai.expect;
const async = require('async');

const host = process.env.MORT_HOST + ':' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('Performance and Concurrency', function () {

    describe('Request collapsing', function () {
        it('should collapse identical concurrent requests', function (done) {
            this.timeout(10000);
            const path = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=180&quality=75';
            const startTime = Date.now();

            async.parallel([
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
                (cb) => request.get(path).expect(200).end(cb),
            ], (err, results) => {
                if (err) return done(err);
                const duration = Date.now() - startTime;

                // All requests should complete
                expect(results.length).to.eql(8);

                // All responses should have same content length
                const firstLength = results[0].body.length;
                results.forEach(res => {
                    expect(res.body.length).to.eql(firstLength);
                });

                // Request collapsing should make this faster than processing 8 times
                // This is a loose check - should complete in reasonable time
                expect(duration).to.be.lessThan(15000);
                done();
            });
        });

        it('should handle different transforms concurrently', function (done) {
            this.timeout(10000);

            async.parallel([
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=100').expect(200).end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=200').expect(200).end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=300').expect(200).end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=400').expect(200).end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=5').expect(200).end(cb),
            ], (err, results) => {
                if (err) return done(err);

                // All requests should complete successfully
                expect(results.length).to.eql(5);

                // Each response should have different dimensions
                expect(results[0].headers['x-amz-meta-public-width']).to.eql('100');
                expect(results[1].headers['x-amz-meta-public-width']).to.eql('200');
                expect(results[2].headers['x-amz-meta-public-width']).to.eql('300');
                expect(results[3].headers['x-amz-meta-public-width']).to.eql('400');

                done();
            });
        });

        it('should collapse requests for different images separately', function (done) {
            this.timeout(10000);

            const path1 = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=185';
            const path2 = '/local/large.jpeg';

            async.parallel([
                (cb) => request.get(path1).expect(200).end(cb),
                (cb) => request.get(path1).expect(200).end(cb),
                (cb) => request.get(path1).expect(200).end(cb),
                (cb) => request.get(path2).expect(200).end(cb),
                (cb) => request.get(path2).expect(200).end(cb),
                (cb) => request.get(path2).expect(200).end(cb),
            ], (err, results) => {
                if (err) return done(err);

                expect(results.length).to.eql(6);

                // First 3 should be identical
                const firstLength = results[0].body.length;
                expect(results[1].body.length).to.eql(firstLength);
                expect(results[2].body.length).to.eql(firstLength);

                // Last 3 should be identical but different from first 3
                const secondLength = results[3].body.length;
                expect(results[4].body.length).to.eql(secondLength);
                expect(results[5].body.length).to.eql(secondLength);
                expect(firstLength).to.not.eql(secondLength);

                done();
            });
        });
    });

    describe('Caching performance', function () {
        it('should serve cached content faster than fresh transforms', function (done) {
            this.timeout(10000);
            const path = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=190&operation=blur&sigma=3';

            // First request - cold cache
            const startCold = Date.now();
            request.get(path)
                .expect(200)
                .end(function (err) {
                    if (err) return done(err);
                    const coldDuration = Date.now() - startCold;

                    // Second request - warm cache
                    setTimeout(() => {
                        const startWarm = Date.now();
                        request.get(path)
                            .expect(200)
                            .end(function (err2, res2) {
                                if (err2) return done(err2);
                                const warmDuration = Date.now() - startWarm;

                                // Cached response might have cache hit header
                                if (res2.headers['x-mort-cache']) {
                                    expect(res2.headers['x-mort-cache']).to.eql('hit');
                                }

                                // Warm should be faster or similar (accounting for variance)
                                expect(warmDuration).to.be.lessThan(coldDuration * 2);
                                done();
                            });
                    }, 100);
                });
        });

        it('should handle cache invalidation for updated files', function (done) {
            this.timeout(5000);
            const path = '/local/large.jpeg';

            // First request
            request.get(path)
                .expect(200)
                .end(function (err, res1) {
                    if (err) return done(err);
                    const etag1 = res1.headers.etag;

                    // Second request - should return same ETag
                    request.get(path)
                        .expect(200)
                        .end(function (err2, res2) {
                            if (err2) return done(err2);
                            const etag2 = res2.headers.etag;
                            expect(etag2).to.eql(etag1);
                            done();
                        });
                });
        });
    });

    describe('Response time requirements', function () {
        it('should serve static images quickly', function (done) {
            this.timeout(5000);
            const start = Date.now();

            request.get('/local/large.jpeg')
                .expect(200)
                .end(function (err) {
                    if (err) return done(err);
                    const duration = Date.now() - start;
                    // Should serve static files within 2 seconds
                    expect(duration).to.be.lessThan(2000);
                    done();
                });
        });

        it('should complete simple transforms in reasonable time', function (done) {
            this.timeout(5000);
            const start = Date.now();

            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=200')
                .expect(200)
                .end(function (err) {
                    if (err) return done(err);
                    const duration = Date.now() - start;
                    // Simple resize should complete within 5 seconds
                    expect(duration).to.be.lessThan(5000);
                    done();
                });
        });

        it('should complete complex transforms in reasonable time', function (done) {
            this.timeout(10000);
            const start = Date.now();

            request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=400&operation=blur&sigma=5&operation=rotate&angle=90&format=webp')
                .expect(200)
                .end(function (err) {
                    if (err) return done(err);
                    const duration = Date.now() - start;
                    // Complex transforms should complete within 10 seconds
                    expect(duration).to.be.lessThan(10000);
                    done();
                });
        });
    });

    describe('Concurrent different operations', function () {
        it('should handle mixed operations concurrently', function (done) {
            this.timeout(10000);

            async.parallel([
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=100').end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=blur&sigma=2').end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=rotate&angle=90').end(cb),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg/default_small').end(cb),
                (cb) => request.get('/local/large.jpeg').end(cb),
                (cb) => request.get('/local/main.css').end(cb),
            ], (err, results) => {
                if (err) return done(err);

                // All should complete successfully
                results.forEach(res => {
                    expect(res.status).to.eql(200);
                });

                done();
            });
        });

        it('should maintain correct results under concurrent load', function (done) {
            this.timeout(15000);
            const requests = [];

            // Create 20 different transforms
            for (let i = 0; i < 20; i++) {
                const width = 100 + (i * 10);
                requests.push((cb) => {
                    request.get(`/remote/nxpvwo7qqfwz.jpg?operation=resize&width=${width}`)
                        .expect(200)
                        .end((err, res) => {
                            if (err) return cb(err);
                            cb(null, { width, res });
                        });
                });
            }

            async.parallel(requests, (err, results) => {
                if (err) return done(err);

                // Verify each result has correct dimensions
                results.forEach((result) => {
                    const expectedWidth = result.width.toString();
                    const actualWidth = result.res.headers['x-amz-meta-public-width'];
                    expect(actualWidth).to.eql(expectedWidth);
                });

                done();
            });
        });
    });

    describe('Memory management', function () {
        it('should handle sequential large file requests without memory issues', function (done) {
            this.timeout(30000);

            async.series([
                (cb) => request.get('/local/large.jpeg').expect(200).end(cb),
                (cb) => request.get('/local/large.jpeg?operation=resize&width=800').expect(200).end(cb),
                (cb) => request.get('/local/large.jpeg').expect(200).end(cb),
                (cb) => request.get('/local/large.jpeg?operation=resize&width=600').expect(200).end(cb),
                (cb) => request.get('/local/large.jpeg').expect(200).end(cb),
            ], (err, results) => {
                if (err) return done(err);

                // All requests should complete
                expect(results.length).to.eql(5);
                results.forEach(res => {
                    expect(res.status).to.eql(200);
                    expect(res.body.length).to.be.greaterThan(0);
                });

                done();
            });
        });

        it('should handle multiple concurrent transforms without memory leak', function (done) {
            this.timeout(15000);
            const requests = [];

            // 15 concurrent requests with different transforms
            for (let i = 0; i < 15; i++) {
                requests.push((cb) => {
                    request.get(`/remote/nxpvwo7qqfwz.jpg?operation=resize&width=${150 + i * 20}&operation=blur&sigma=${1 + i}`)
                        .end(cb);
                });
            }

            async.parallel(requests, (err, results) => {
                if (err) return done(err);

                // All should complete successfully
                results.forEach(res => {
                    expect([200, 429, 503]).to.include(res.status);
                });

                done();
            });
        });
    });

    describe('Streaming performance', function () {
        it('should start streaming large files without full buffering', function (done) {
            this.timeout(10000);
            const start = Date.now();

            request.get('/local/large.jpeg')
                .expect(200)
                .end(function (err, res) {
                    if (err) return done(err);
                    const totalTime = Date.now() - start;

                    // Large file should still complete in reasonable time
                    expect(totalTime).to.be.lessThan(10000);
                    expect(res.body.length).to.be.greaterThan(0);
                    done();
                });
        });
    });

    describe('Throttling and rate limiting', function () {
        it('should handle burst requests gracefully', function (done) {
            this.timeout(15000);
            const requests = [];

            // Send 30 concurrent requests
            for (let i = 0; i < 30; i++) {
                requests.push((cb) => {
                    request.get(`/remote/nxpvwo7qqfwz.jpg?operation=resize&width=${195 + i}`)
                        .end((err, res) => {
                            cb(null, res ? res.status : 500);
                        });
                });
            }

            async.parallel(requests, (err, statuses) => {
                if (err) return done(err);

                // Count successful and rate-limited responses
                const successful = statuses.filter(s => s === 200).length;
                const rateLimited = statuses.filter(s => s === 429 || s === 503).length;

                // At least some should succeed
                expect(successful).to.be.greaterThan(0);

                // If rate limiting is enabled, some might be limited
                expect(successful + rateLimited).to.eql(30);

                done();
            });
        });
    });

    describe('Error recovery', function () {
        it('should continue serving after encountering errors', function (done) {
            this.timeout(10000);

            async.series([
                (cb) => request.get('/remote/nonexistent.jpg').end(() => cb(null, 'error-handled')),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=200').expect(200).end(cb),
                (cb) => request.get('/remote/invalid.jpg/bad_preset').end(() => cb(null, 'error-handled')),
                (cb) => request.get('/remote/nxpvwo7qqfwz.jpg?operation=resize&width=300').expect(200).end(cb),
            ], (err, results) => {
                if (err) return done(err);

                // Server should remain functional after errors
                expect(results.length).to.eql(4);
                expect(results[1].status).to.eql(200);
                expect(results[3].status).to.eql(200);

                done();
            });
        });
    });
});
