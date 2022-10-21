const supertest  = require('supertest');
const chai = require('chai');
const expect = chai.expect;
const async = require('async')

const host = process.env.MORT_HOST + ':' + + process.env.MORT_PORT;
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


describe('Image processing - cache', function () {

    it('should create thumbnails and store in cache', function (done) {
        this.timeout(5000);
        async.Seriesk
        const reqPath = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=300&operation=rotate&angle=90&format=webp';
        const width = '300';
        const height = '240';
        const sizeRange = [5100, 8100]
        async.parallel([
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
            (cb)  => checkImage(reqPath, width, height, sizeRange, cb),
        ], (err) => {
            if (err) {
                return done(err)
            }
            request.get(reqPath)
                .expect(200)
                .end(function(err2, res) {
                    if (err) {
                        return done(err2);
                    }

                    expect(res.headers['x-mort-cache']).to.eql('hit');
                    done(err)
                })
        })
    });

    it('should create thumbnails with watermark and cache', function (done) {
        this.timeout(5000);
        const reqPath = '/remote/nxpvwo7qqfwz.jpg?operation=resize&width=400&height=200&image=https://i.imgur.com/uomkVIL.png&opacity=0.5&position=top-left&operation=watermark';
        const width = '400';
        const height = '200';
        async.parallel([
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
            (cb)  => checkImage(reqPath, width, height, [5645, 18300], cb),
        ], (err) => {
            if (err) {
                return done(err)
            }
            request.get(reqPath)
                .expect(200)
                .end(function(err2, res) {
                    if (err) {
                        return done(err2);
                    }

                    expect(res.headers['x-mort-cache']).to.eql('hit');
                    done(err)
                })
        })
    });

});