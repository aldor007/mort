const axiosRetry = require('axios-retry');
const axios = require('axios');

axiosRetry(axios, { retries: 10,
    retryDelay: axiosRetry.exponentialDelay,
    retryCondition: (err) => {
        if (err.response && err.response.status > 0) {
            return false
        }

        return true
    }
});


before(async function () {  
    this.timeout(60000)
    try  {
        await axios.get(`http://${process.env.MORT_HOST}:${process.env.MORT_PORT}`)
    } catch (e) {

    }
})