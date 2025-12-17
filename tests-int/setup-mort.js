const axiosRetry = require('axios-retry').default;
const axios = require('axios');

// Configure axios with retry for health checks
axiosRetry(axios, {
    retries: 10,
    retryDelay: axiosRetry.exponentialDelay,
    retryCondition: (err) => {
        if (err.response && err.response.status > 0) {
            return false
        }
        return true
    }
});

// Global retry configuration for tests
global.TEST_RETRY_CONFIG = {
    maxRetries: 3,
    retryDelay: 1000,
    shouldRetry: function(err, res) {
        // Retry on connection errors
        if (err && (
            err.code === 'ECONNREFUSED' ||
            err.code === 'ECONNRESET' ||
            err.code === 'ETIMEDOUT' ||
            err.code === 'ENOTFOUND' ||
            err.code === 'ENETUNREACH' ||
            err.code === 'EPIPE'
        )) {
            return true;
        }

        // Retry on 5xx errors except 501 (Not Implemented)
        if (res && res.status >= 500 && res.status !== 501) {
            return true;
        }

        return false;
    }
};

// Helper function to wrap test requests with retry logic
global.withRetry = function(requestFn, options = {}) {
    const maxRetries = options.maxRetries || global.TEST_RETRY_CONFIG.maxRetries;
    const retryDelay = options.retryDelay || global.TEST_RETRY_CONFIG.retryDelay;

    return function(callback) {
        let attempt = 0;

        const tryRequest = () => {
            requestFn((err, res) => {
                if (err || (res && res.error)) {
                    const shouldRetry = global.TEST_RETRY_CONFIG.shouldRetry(err, res);

                    if (shouldRetry && attempt < maxRetries) {
                        attempt++;
                        const delay = retryDelay * Math.pow(2, attempt - 1);
                        console.log(`  → Retrying request (attempt ${attempt}/${maxRetries}) after ${delay}ms`);
                        setTimeout(tryRequest, delay);
                        return;
                    }
                }

                callback(err, res);
            });
        };

        tryRequest();
    };
};

before(async function () {
    this.timeout(60000);
    console.log('Waiting for mort server to be ready...');

    // Wait for server with retries
    for (let i = 0; i < 30; i++) {
        try {
            await axios.get(`http://${process.env.MORT_HOST}:${process.env.MORT_PORT}`, { timeout: 2000 });
            console.log('✓ Mort server is ready!');
            return;
        } catch (e) {
            if (i < 29) {
                await new Promise(resolve => setTimeout(resolve, 1000));
            }
        }
    }
    console.warn('⚠ Mort server may not be fully ready, proceeding anyway...');
});

after(function() {
    console.log('\n✓ Test suite completed');
});
