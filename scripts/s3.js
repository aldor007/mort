const AWS = require('aws-sdk');

const s3opts = {
    region: 'mort',
    endpoint: 'localhost:8080',
    s3ForcePathStyle: true,
    sslEnabled: false,
    accessKeyId: 'acc',
    secretAccessKey: 'sec',
    signatureVersion: 's3',
    computeChecksums: true
};


const s3 = new AWS.S3(s3opts);

const listParams = {
    Bucket: 'assets',
    Prefix: 'css/',
};

s3.listObjects(listParams, function (err, data) {
    console.info(err, data)
});
