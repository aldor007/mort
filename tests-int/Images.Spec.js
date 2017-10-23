const supertest  = require('supertest');

const host = 'localhost:' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);


describe('Image processing', function () {
    it('should create thumbnails with format default_small from external source', function (done) {
       request.get('/remote/ChzUb.jpg/default_small')
           .expect('Content-Type', 'image/jpeg')
           .expect(200)
           .end(function(err, res) {
               done(err)
           });
    });

    it('should return 400 when invalid preset given', function (done) {
        request.get('/remote/ChzUb.jpg/default_smalaaal')
            .expect(400)
            .end(function(err, res) {
                done(err)
            });
    });

    it('should return 404 when parent not found', function (done) {
        request.get('/remote/ChzUbk.jpg/default_small')
            .expect(404)
            .end(function(err, res) {
                done(err)
            });
    });
});