const supertest  = require('supertest');
const chai = require('chai');
const expect = chai.expect;

const host = 'localhost:' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('Image processing', function () {
    it('should create thumbnails with format default_small from external source', function (done) {
       request.get('/remote/ChzUb.jpg/default_small')
           .expect('Content-Type', 'image/jpeg')
           .expect(200)
           .end(function(err, res) {
               expect(res.headers['x-amz-public-width']).to.eql('150');
               expect(res.headers['x-amz-public-height']).to.eql('200');
               request.get('/remote/ChzUb.jpg/default_small')
                   .end(function (err2, res2) {
                       expect(res2.headers['x-amz-public-width']).to.eql('150');
                       expect(res2.headers['x-amz-public-height']).to.eql('200');
                       done(err2)
                   });
           });
    });

    it('should return 400 when invalid preset given', function (done) {
        request.get('/remote/ChzUb.jpg/default_smalaaal')
            .expect(400)
            .end(function(err) {
                done(err)
            });
    });

    it('should return 404 when parent not found', function (done) {
        request.get('/remote/ChzUbk.jpg/default_small')
            .expect(404)
            .end(function(err) {
                done(err)
            });
    });
});