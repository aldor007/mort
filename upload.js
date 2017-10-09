const fs = require('fs')
const AWS = require('aws-sdk');
// const mime = require('mime')

const host = 'localhost:8080'
const accessKeyId = 'acc'
const secretAccessKey = 'sec'

const s3opts = {
   region: 'mort',
   endpoint: host,
   s3ForcePathStyle: true,
   sslEnabled: false,
   accessKeyId: accessKeyId,
   secretAccessKey: secretAccessKey,
   signatureVersion: 's3',
   computeChecksums: true
};
const body = fs.readFileSync('file.jpeg')
const s3 = new AWS.S3(s3opts);

const headers = {}
headers['content-length'] = headers['content-length'] || body.length;
headers['content-type'] = headers['content-type'] ||  'image/jpeg'

const params = {
    Body: body,
    Bucket: 'media',
    Key: '/file.jpg',
    ContentDisposition: headers['content-disposition'],
    ContentEncoding: headers['content-encoding'],
    ContentLanguage: headers['content-language'],
    ContentType: headers['content-type'],
    ContentLength: headers['content-length'],
    Etag: headers['etag'],
    Metadata: {}
};

const options = {
    partSize: 50 * 1024 * 1024,
    queueSize: 1
};

s3.listBuckets({}, function (err, data) {
	console.info(err, data)
})
// s3.upload(params, options, function (err, data) {
//     if (err) {
//         console.error('Error uploading file', err);
//     }

//     console.info('Successful uploaded file');
// });
