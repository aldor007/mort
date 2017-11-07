const chai = require('chai');
const expect = chai.expect;
const AWS = require('aws-sdk');
const supertest  = require('supertest');

const host = 'localhost:' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('S3 features', function () {
    beforeEach(function () {
       this.s3opts = {
            region: 'mort',
            endpoint: 'localhost:' + process.env.MORT_PORT,
            s3ForcePathStyle: true,
            sslEnabled: false,
            accessKeyId: 'acc',
            secretAccessKey: 'sec',
            signatureVersion: 's3',
            computeChecksums: true
        };
    });

    describe('auth s3', function () {
        beforeEach(function () {
            this.s3opts.signatureVersion = 's3';
            this.s3opts.accessKeyId = 'acc';
            this.s3 = new AWS.S3(this.s3opts);
        });

        describe('list buckets', function () {
            it('should list buckets', function (done) {
                this.s3.listBuckets({}, function (err, data) {
                    expect(err).to.be.null;
                    expect(Object.keys(data['Buckets']).length).to.eql(2);
                    expect(data['Buckets'][0].Name).to.eql('local');
                    expect(data['Buckets'][1].Name).to.eql('remote');
                    done(err)
                })
            });

            it('should return error when listBuckets with invalid accessKey', function (done) {
                this.s3opts.accessKeyId = 'invalid';
                this.s3 = new AWS.S3(this.s3opts);
                const listParams = {
                    Bucket: 'local'
                };

                this.s3.listBuckets({}, function (err, data) {
                    expect(data).to.be.null;
                    expect(err).to.be.an('error');
                    done()
                });
            });
        });

        describe('list files', function () {
            it('should list files', function (done) {
                const listParams = {
                    Bucket: 'local'
                };

                this.s3.listObjects(listParams, function (err, data) {
                    expect(err).to.be.null;
                    expect(data['CommonPrefixes']).to.deep.eql([ { Prefix: 'dir/' } ]);
                    expect(data['Contents'].length).to.eql(1);
                    done(err)
                });
            });

            it('should return error when listObject with invalid accessKey', function (done) {
                this.s3opts.accessKeyId = 'invalid';
                this.s3 = new AWS.S3(this.s3opts);
                const listParams = {
                    Bucket: 'local'
                };

                this.s3.listObjects(listParams, function (err, data) {
                    expect(err).to.be.an('error');
                    expect(data).to.be.null;
                    done()
                });
            });
        });

        describe('uploading file', function () {
            it('should upload file', function (done) {
                const headers = {};
                const body = 'aaaa body';
                headers['content-length'] = headers['content-length'] || body.length;
                headers['content-type'] = headers['content-type'] ||  'image/jpeg'

                const params = {
                    Body: body,
                    Bucket: 'local',
                    Key: 'file.jpg',
                    ContentType: headers['content-type'],
                    ContentLength: headers['content-length'],
                    Etag: headers['etag'],
                    Metadata: {}
                };

                this.s3.upload(params, function (err, data) {
                    expect(err).to.be.null;
                    done(err)
                });
            });

            it('should return error when invalid access key', function (done) {
                this.s3opts.accessKeyId = 'invalid';
                this.s3 = new AWS.S3(this.s3opts);
                const headers = {};
                const body = 'aaaa body';
                headers['content-length'] = headers['content-length'] || body.length;
                headers['content-type'] = headers['content-type'] ||  'image/jpeg';

                const params = {
                    Body: body,
                    Bucket: 'local',
                    Key: 'file.jpg',
                    ContentType: headers['content-type'],
                    ContentLength: headers['content-length'],
                    Etag: headers['etag'],
                    Metadata: {}
                };

                this.s3.upload(params, function (err, data) {
                    expect(err).to.be.an('error');
                    done()
                });
            });
        });
    });

    describe('auth v4', function () {
        beforeEach(function () {
            this.s3opts.signatureVersion = 'v4';
            this.s3opts.accessKeyId = 'acc';
            this.s3 = new AWS.S3(this.s3opts);
        });

        describe('list buckets', function () {
            it('should list buckets', function (done) {
                this.s3.listBuckets({}, function (err, data) {
                    expect(err).to.be.null;
                    expect(Object.keys(data['Buckets']).length).to.eql(2);
                    expect(data['Buckets'][0].Name).to.eql('local');
                    expect(data['Buckets'][1].Name).to.eql('remote');
                    done(err)
                })
            });

            it('should return error when listBuckets with invalid accessKey', function (done) {
                this.s3opts.accessKeyId = 'invalid';
                this.s3 = new AWS.S3(this.s3opts);
                const listParams = {
                    Bucket: 'local'
                };

                this.s3.listBuckets({}, function (err, data) {
                    expect(err).to.be.an('error');
                    expect(data).to.be.null;
                    done()
                });
            });
        });

        describe('list files', function () {
            it('should list files', function (done) {
                const listParams = {
                    Bucket: 'local'
                };

                this.s3.listObjects(listParams, function (err, data) {
                    expect(err).to.be.null;
                    expect(data['CommonPrefixes']).to.deep.eql([ { Prefix: 'dir/' } ]);
                    expect(data['Contents'].length).to.eql(2);
                    done(err)
                });
            });

            it('should return error when listObject with invalid accessKey', function (done) {
                this.s3opts.accessKeyId = 'invalid';
                this.s3 = new AWS.S3(this.s3opts);
                const listParams = {
                    Bucket: 'local'
                };

                this.s3.listObjects(listParams, function (err, data) {
                    expect(err).to.be.an('error');
                    expect(data).to.be.null;
                    done()
                });
            });
        });

        describe('uploading file', function () {
            it('should upload file', function (done) {
                const headers = {};
                const body = 'aaaa body';
                headers['content-length'] = headers['content-length'] || body.length;
                headers['content-type'] = headers['content-type'] ||  'image/jpeg';

                const params = {
                    Body: body,
                    Bucket: 'local',
                    Key: 'file.jpg',
                    ContentType: headers['content-type'],
                    ContentLength: headers['content-length'],
                    Etag: headers['etag'],
                    Metadata: {
                        'header': 'meta'
                    }
                };

                this.s3.upload(params, function (err, data) {
                    expect(err).to.be.null;
                    done(err)
                });
            });

            it('should return valid metadata for uploaded file', function (done) {
                request.get('/local/file.jpg')
                    .expect(200)
                    .end(function(err, res) {
                        expect(res.headers['x-amz-meta-header']).to.eql('meta');
                        done(err)
                    });
            });

            it('should return error when invalid access key', function (done) {
                this.s3opts.accessKeyId = 'invalid';
                this.s3 = new AWS.S3(this.s3opts);
                const headers = {};
                const body = 'aaaa body';
                headers['content-length'] = headers['content-length'] || body.length;
                headers['content-type'] = headers['content-type'] ||  'image/jpeg';

                const params = {
                    Body: body,
                    Bucket: 'local',
                    Key: 'file.jpg',
                    ContentType: headers['content-type'],
                    ContentLength: headers['content-length'],
                    Etag: headers['etag'],
                    Metadata: {}
                };

                this.s3.upload(params, function (err, data) {
                    expect(err).to.be.an('error');
                    done()
                });
            });
        });
    });

});
