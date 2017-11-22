const supertest  = require('supertest');
const chai = require('chai');
const expect = chai.expect;

const host = 'localhost:' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('Image processing', function () {
    describe('presets', () => {
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
                    expect(body.length).to.be.within(9300, 9600);

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

        it('should return 400 when invalid preset given', function (done) {
            request.get('/remote/nxpvwo7qqfwz.jpg/default_smalaaal')
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
        it('should create thumbnails with format with = 100 and format webpfrom external source', function (done) {
            this.timeout(5000);
            const reqPath = '/remote-query/nxpvwo7qqfwz.jpg?width=100&format=webp';
            request.get(reqPath)
                .expect(200)
                .expect('Content-Type', 'image/webp')
                .end(function(err, res) {
                    if (err) {
                        return done(err);
                    }

                    const body = res.body;
                    expect(body.length).to.be.within(1900, 2000);

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

        it('should return 404 when parent not found', function (done) {
            request.get('/remote-query/path/a.png')
                .expect(404)
                .end(function(err) {
                    done(err)
                });
        });
    });
});