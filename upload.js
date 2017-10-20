const fs = require('fs')
const AWS = require('aws-sdk');
// const mime = require('mime')

const host = 'localhost:8080'
// const host = 'mort.mkaciuba.com'
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
    Bucket: 'media2',
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

// s3.listBuckets({}, function (err, data) {
// 	console.info(err, data)
// })

const listParams = {
    Bucket: 'liip',
}

// s3.listObjects(listParams, function (err, data) {
//     if (err) {
//         console.error(err);
//         throw err;
//     }
//     console.info('list', data)
// })

s3.upload(params, options, function (err, data) {
    if (err) {
        console.error('Error uploading file', err);
        return;
    }

    console.info('Successful uploaded file');
});
