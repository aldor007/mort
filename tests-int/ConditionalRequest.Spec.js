const supertest  = require('supertest');
const moment = require('moment');

const host = process.env.MORT_HOST + ':' + + process.env.MORT_PORT;
const request = supertest(`http://${host}`);
const filePath = '/local/large.jpeg';

describe('HTTP conditional requests',  function () {
    let etag = null;
    let lastMod = null;

    before(function (done) {
       request.head(filePath)
           .end(function (_, res) {
               etag = res.headers['etag'];
               lastMod = res.headers['last-modified'];
               done();
           })
    });

    it('should return 304 for if-none-match', function (done) {
        request.get(filePath)
            .set('if-none-match', etag)
            .expect(304)
            .end(done);
    });

    it('should return 304 for if-modified-since', function (done) {
        request.get(filePath)
            .set('If-modified-Since', lastMod)
            .expect(304)
            .end(done);
    });
});
