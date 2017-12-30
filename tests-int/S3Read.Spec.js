const chai = require('chai');
const expect = chai.expect;
const AWS = require('aws-sdk');
const supertest  = require('supertest');

const host = 'localhost:' + process.env.MORT_PORT;
const request = supertest(`http://${host}`);

describe('S3 Read features', function () {
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

        this.s3opts.signatureVersion = 's3';
        this.s3opts.accessKeyId = 'acc';
        this.s3 = new AWS.S3(this.s3opts);
    });

    describe('head bucket', function () {

        it('bucket local should exists', function (done) {
            this.s3.headBucket({
                Bucket: 'local'
            }, function (err, data) {
                if (err) {
                   return done(err)
                }

                expect(data).not.to.be.null;
                done();
            })
        });

        it('bucket local2 shouldn\'t exists', function (done) {
            this.s3.headBucket({
                Bucket: 'local2'
            }, function (err, data) {
                expect(err).not.to.be.null;
                done();
            })
        });

    });

    describe('head and create dir', function () {
        it('should return error when dir doesn\'t exist', function (done) {
           const params = {
               Bucket: 'local',
               Key: 'dir-2'
           };

            this.s3.headObject(params, function (err) {
                expect(err).not.to.be.null;
                done();
            })
        });

       it('should create dir', function (done) {
           const params = {
               Bucket: 'local',
               Key: 'dir-11/',
               Body: ''
           };

           this.s3.upload(params, function (err) {
               expect(err).to.be.null;
               done();
           })
        });

    })
});
