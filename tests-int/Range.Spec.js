const fs = require('fs');
const binary = require('superagent-binary-parser');

const file = fs.readFileSync('./pkg/processor/benchmark/local/large.jpeg');
const supertest  = require('supertest');

const host = process.env.MORT_HOST + ':' + + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

function getRange(range) {
    range = range.split('=')[1].split('-');
    range = range.map(function (r) {
        return parseInt(r, 10);
    });
    if (range[1] >= file.length) {
        range[1] = file.length - 1;
    }
    return range;
}

describe('Range', function () {

    it('should handle single range', function (done) {
        request.get('/local/large.jpeg')
            .set('Range', 'bytes=0-1000')
            .parse(binary)
            .buffer()
            .expect('Content-Range', 'bytes 0-1000/' + file.length)
            .expect(206, file.slice(0, 1001))
            .end(done);
    });

    it('should handle multiple ranges', function (done) {

        request.get('/local/large.jpeg')
            .set('Range', 'bytes=2000-4000,6000-6500')
            .parse(binary)
            .buffer()
            .expect(function (res) {
                var str = res.body.toString('utf-8');
                if (str.indexOf('Content-Range: bytes 2000-4000/' + file.length) === -1) {
                    throw new Error('Missing first range');
                }
                if (str.indexOf('Content-Range: bytes 6000-6500/' + file.length) === -1) {
                    throw new Error('Missing second range');
                }
                if (res.body.length <= 2502) { // first range + second range (+ delimiters)
                    throw new Error('Response too small');
                }
            })
            .expect(206)
            .end(done);
    });

    it('should handle partial ranges', function (done) {
        request.get('/local/large.jpeg')
            .set('Range', 'bytes=20000-')
            .parse(binary)
            .buffer()
            .expect('Content-Range', 'bytes 20000-' + (file.length - 1) + '/' + file.length)
            .expect(206, file.slice(20000))
            .end(done);
    });


    it('should return 416  on unparsable range', function (done) {
        request.get('/local/large.jpeg')
            .set('Range', 'litres=3.14-1.68')
            .parse(binary)
            .buffer()
            .expect(416)
            .end(done);
    });

});
