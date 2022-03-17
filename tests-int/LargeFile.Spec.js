
const request = require('superagent');
const chai = require('chai');
const expect = chai.expect;
const fs = require('fs');
const crypto = require('crypto');


const url = 'http://localhost:' + process.env.MORT_PORT;


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


describe('Large file', async () =>  {
    it('should download 1GB file', async () => {
        const reqPath = '/local/big.img'
        const expectHash = await hashFile(fs.createReadStream('/tmp/mort-tests/local/big.img'))
        const res = request.get(url + reqPath)
        const hashReq = await hashFile(res);
        expect(hashReq).to.eql(expectHash)
    }).timeout(60000)
})