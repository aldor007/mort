const supertest = require('supertest');

/**
 * Wrapper around supertest that adds retry logic for flaky tests
 */
class RetryableRequest {
    constructor(baseUrl, maxRetries = 3, retryDelay = 1000) {
        this.request = supertest(baseUrl);
        this.maxRetries = maxRetries;
        this.retryDelay = retryDelay;
    }

    /**
     * Execute a request with retry logic
     * @param {string} method - HTTP method (get, post, put, delete, etc.)
     * @param {string} path - Request path
     * @param {object} options - Additional options (timeout, etc.)
     * @returns {Promise} - Supertest Test object
     */
    async executeWithRetry(method, path, options = {}) {
        let lastError;
        const maxRetries = options.maxRetries || this.maxRetries;
        const retryDelay = options.retryDelay || this.retryDelay;

        for (let attempt = 0; attempt <= maxRetries; attempt++) {
            try {
                const req = this.request[method](path);

                // Apply timeout if specified
                if (options.timeout) {
                    req.timeout(options.timeout);
                }

                return req;
            } catch (error) {
                lastError = error;

                // Check if error is retryable
                if (this.isRetryableError(error) && attempt < maxRetries) {
                    // Wait before retrying with exponential backoff
                    const delay = retryDelay * Math.pow(2, attempt);
                    await this.sleep(delay);
                    continue;
                }

                // If not retryable or max retries reached, throw
                throw error;
            }
        }

        throw lastError;
    }

    /**
     * Determine if an error is retryable
     */
    isRetryableError(error) {
        if (!error) return false;

        // Retry on connection errors
        if (error.code === 'ECONNREFUSED' ||
            error.code === 'ECONNRESET' ||
            error.code === 'ETIMEDOUT' ||
            error.code === 'ENOTFOUND' ||
            error.code === 'ENETUNREACH') {
            return true;
        }

        // Retry on 5xx errors except 501 (Not Implemented)
        if (error.status >= 500 && error.status !== 501) {
            return true;
        }

        return false;
    }

    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Convenience methods
    get(path, options = {}) {
        return this.request.get(path);
    }

    post(path, options = {}) {
        return this.request.post(path);
    }

    put(path, options = {}) {
        return this.request.put(path);
    }

    delete(path, options = {}) {
        return this.request.delete(path);
    }

    head(path, options = {}) {
        return this.request.head(path);
    }

    patch(path, options = {}) {
        return this.request.patch(path);
    }
}

/**
 * Create a retryable request instance
 */
function createRetryableRequest(baseUrl, maxRetries = 3, retryDelay = 1000) {
    return new RetryableRequest(baseUrl, maxRetries, retryDelay);
}

module.exports = {
    RetryableRequest,
    createRetryableRequest
};
