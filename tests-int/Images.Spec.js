const supertest  = require('supertest');
const chai = require('chai');
const expect = chai.expect;

const host = 'localhost:' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

function checkImage(reqPath, width, height, sizeRange, done) {
    request.get(reqPath)
        .expect(200)
        .end(function(err, res) {
            if (err) {
                return done(err);
            }

            const body = res.body;
            expect(body.length).to.be.within(...sizeRange);

            expect(res.headers['x-amz-meta-public-width']).to.eql(width);
            expect(res.headers['x-amz-meta-public-height']).to.eql(height);
            request.get(reqPath)
                .expect(200)
                .end(function (err2, res2) {
                    if (err2) {
                        return done(err2);
                    }


                    expect(res2.headers['x-amz-meta-public-width']).to.eql(width);
                    expect(res2.headers['x-amz-meta-public-height']).to.eql(height);

                    const body2 = res2.body;
                    expect(body2.length).to.eql(body.length);
                    done(err2)
                });
        });
}


describe('Image processing', function () {
    describe('presets', function () {

        it('should create thumbnails with format default_small from external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg/default_small';
            request.get(reqPath)
                .expect('Content-Type', 'image/webp')
                .expect(200)
                .end(function(err, res) {
                    if (err) {
                        return done(err);
                    }

                    const body = res.body;
                    expect(body.length).to.be.within(9300, 10000);

                    expect(res.headers['x-amz-meta-public-width']).to.eql('150');
                    expect(res.headers['x-amz-meta-public-height']).to.eql('200');
                    request.get(reqPath)
                        .expect(200)
                        .end(function (err2, res2) {
                            if (err2) {
                                return done(err2);
                            }

                            const body2 = res2.body;
                            expect(body2.length).to.eql(body.length);

                            expect(res2.headers['x-amz-meta-public-width']).to.eql('150');
                            expect(res2.headers['x-amz-meta-public-height']).to.eql('200');
                            done(err2)
                        });
                });
        });

        it('should create thumbnails with crop from external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg/crop';
            const width = '756';
            const height = '396';
            checkImage(reqPath, width, height, [30000, 33000], done)
        });

        it('should create thumbnails with blur from external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg/blur';
            const width = '756';
            const height = '396';
            checkImage(reqPath, width, height, [10222, 11000], done);
        });

        it('should create thumbnails with watermark from external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg/watermark';
            const width = '200';
            const height = '200';
            checkImage(reqPath, width, height, [3500, 4600], done)
        });

        it('should extract given area from image', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg/extract';
            const width = '700';
            const height = '496';
            checkImage(reqPath, width, height, [60000, 70855], done)
        });

        it('should create thumbnails with watermark from external source and handle accept header', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg/watermark';
            const width = '200';
            const height = '200';
            request.get(reqPath)
                .set('accept', 'image/webp')
                .expect(200)
                .end(function(err, res) {
                    if (err) {
                        return done(err);
                    }

                    const body = res.body;
                    expect(body.length).to.be.within(2000, 2700);

                    expect(res.headers['x-amz-meta-public-width']).to.eql(width);
                    expect(res.headers['x-amz-meta-public-height']).to.eql(height);
                    expect(res.headers['content-type']).to.eql('image/webp');
                    request.get(reqPath)
                        .expect(200)
                        .set('accept', 'image/webp')
                        .end(function (err2, res2) {
                            if (err2) {
                                return done(err2);
                            }

                            const body2 = res2.body;
                            expect(body2.length).to.eql(body.length);

                            expect(res2.headers['x-amz-meta-public-width']).to.eql(width);
                            expect(res2.headers['x-amz-meta-public-height']).to.eql(height);
                            expect(res2.headers['content-type']).to.eql('image/webp');
                            done(err2)
                        });
                });
        });

        it('should return 400 when invalid preset given', function (done) {
            request.get('/local/nxpvwo7qqfwz.jpg/default_smalaaal')
                .expect(400)
                .end(function(err) {
                    done(err)
                });
        });

        it('should return 404 when parent not found', function (done) {
            request.get('/remote/nie.ma/default_small')
                .expect(404)
                .end(function(err) {
                    done(err)
                });
        });
    });

    describe('query', function ()  {
        it('should create thumbnails with width = 100 and format webp from external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg?width=100&format=webp';
            request.get(reqPath)
                .expect(200)
                .expect('Content-Type', 'image/webp')
                .end(function(err, res) {
                    if (err) {
                        return done(err);
                    }

                    const body = res.body;
                    expect(body.length).to.be.within(1500, 2000);

                    expect(res.headers['x-amz-meta-public-width']).to.eql('100');
                    expect(res.headers['x-amz-meta-public-height']).to.eql('125');
                    request.get(reqPath)
                        .expect(200)
                        .end(function (err2, res2) {
                            if (err2) {
                                return done(err2);
                            }

                            const body2 = res2.body;
                            expect(body2.length).to.eql(body.length);

                            expect(res2.headers['x-amz-meta-public-width']).to.eql('100');
                            expect(res2.headers['x-amz-meta-public-height']).to.eql('125');
                            done(err2)
                        });
                });
        });

        it('should create thumbnails with format width = 100, height = 100 and format webp from external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg?width=100&format=webp&height=100';
            request.get(reqPath)
                .expect(200)
                .expect('Content-Type', 'image/webp')
                .end(function(err, res) {
                    if (err) {
                        return done(err);
                    }

                    const body = res.body;
                    expect(body.length).to.be.within(1300, 2054);

                    expect(res.headers['x-amz-meta-public-width']).to.eql('100');
                    expect(res.headers['x-amz-meta-public-height']).to.eql('100');
                    request.get(reqPath)
                        .expect(200)
                        .end(function (err2, res2) {
                            if (err2) {
                                return done(err2);
                            }

                            const body2 = res2.body;
                            expect(body2.length).to.eql(body.length);

                            expect(res2.headers['x-amz-meta-public-width']).to.eql('100');
                            expect(res2.headers['x-amz-meta-public-height']).to.eql('100');
                            done(err2)
                        });
                });
        });

        it('should create thumbnails with many operations', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=400&format=webp&height=100&operation=blur&sigma=5&grayscale=1';
            request.get(reqPath)
                .expect(200)
                .expect('Content-Type', 'image/webp')
                .end(function(err, res) {
                    if (err) {
                        return done(err);
                    }

                    const body = res.body;
                    const width = '400';
                    const height = '100';
                    expect(body.length).to.be.within(1060, 2000);

                    expect(res.headers['x-amz-meta-public-width']).to.eql(width);
                    expect(res.headers['x-amz-meta-public-height']).to.eql(height);
                    request.get(reqPath)
                        .expect(200)
                        .end(function (err2, res2) {
                            if (err2) {
                                return done(err2);
                            }

                            const body2 = res2.body;
                            expect(body2.length).to.eql(body.length);

                            expect(res2.headers['x-amz-meta-public-width']).to.eql(width);
                            expect(res2.headers['x-amz-meta-public-height']).to.eql(height);
                            done(err2)
                        });
                });
        });

        it('should create thumbnails and rotate', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=400&operation=rotate&angle=90';
            const width = '400';
            const height = '320';
            checkImage(reqPath, width, height, [10100, 18300], done);
        });

        it('should create thumbnails with watermark', function (done) {
            this.timeout(5000);
            const reqPath = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=400&height=100&image=https://i.imgur.com/uomkVIL.png&opacity=0.5&position=top-left&operation=watermark';
            const width = '400';
            const height = '100';
            checkImage(reqPath, width, height, [3500, 4600], done);
        });

        it('should return 404 when parent not found', function (done) {
            request.get('/remote/a.png')
                .expect(404)
                .end(function(err) {
                    done(err)
                });
        });
    });
});