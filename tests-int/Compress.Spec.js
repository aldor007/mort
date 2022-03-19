
const supertest  = require('supertest');
const crypto = require('crypto');
const { expect } = require('chai');
const fs = require('fs');
const zlib = require('zlib');
const stream = require('stream');
const binary = require('superagent-binary-parser');

const host = process.env.MORT_HOST + ':' + + process.env.MORT_PORT;
const request = supertest(`http://${host}`);
const filePath = '/local/main.css';

const hashFile = async (input) => {
  let hash = crypto.createHash('sha256');
  hash.setEncoding('hex');

  return new Promise((resolve, reject) => {
    input.on('end', () => {
      hash.end();
      let hashHex = hash.read();
      resolve(hashHex);
    });
    input.pipe(hash);
  });
};

const hashBuffer = (data) => {
  return crypto.createHash('sha256').update(data).digest('hex')
}

describe('Compression', function () {
    describe('gzip', function () {
        it('should return gzip response', function (done) {
            request.get(filePath)
                .set('Accept-Encoding', 'gzip')
                .expect(200)
                .expect('Content-Encoding', 'gzip')
                .end(async (err, res) => {
                    if(err) done(err)
                    const expectedHash = await hashFile(fs.createReadStream('/tmp/mort-tests/local/main.css'))
                    expect(hashBuffer(res.text)).to.be.eql(expectedHash)
                    done()
                })
        });
    });
    
    describe('br', function () {
        it('should return br response', function (done) {
            request.get(filePath)
                .set('accept-encoding', 'br, gzip')
                .parse(binary)
                .expect(200)
                .expect('content-encoding', 'br')
                .end(async (err, res) => {
                    if(err) done(err)
                    const expectedHash = await hashFile(fs.createReadStream('/tmp/mort-tests/local/main.css'))
                    zlib.brotliDecompress(res.body, async (err, body) => {
                        if (err) done(err)
                        const hash = await hashFile(stream.Readable.from(body))
                        expect(hash).to.be.eql(expectedHash)
                        done()

                    })

                })
        })
    });
})