import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const errorRate = new Rate('errors');
const authDuration = new Trend('auth_duration');
const infoDuration = new Trend('info_duration');
const sendCoinDuration = new Trend('send_coin_duration');
const buyItemDuration = new Trend('buy_item_duration');

export const options = {
    scenarios: {
        constant_load: {
            executor: 'constant-arrival-rate',
            rate: 1000,
            timeUnit: '1s',
            duration: '5m',
            preAllocatedVUs: 100,
            maxVUs: 1000,
        },
    },
    thresholds: {
        'http_req_duration': ['p(99.99) < 50'],
        'errors': ['rate<0.0001'],
        'auth_duration': ['p(95) < 50'],
        'info_duration': ['p(95) < 50'],
        'send_coin_duration': ['p(95) < 50'],
        'buy_item_duration': ['p(95) < 50'],
    },
};

// Shared data
const BASE_URL = 'http://localhost:8080/api';
const availableItems = [
    't-shirt', 'cup', 'book', 'pen', 'powerbank',
    'hoody', 'umbrella', 'socks', 'wallet', 'pink-hoody'
];

// Test setup - будет выполняться для каждого VU
export function setup() {
    const users = Array.from({ length: 100 }, (_, i) => ({
        username: `testuser${i}`,
        password: 'password123'
    }));
    return { users };
}

export default function(data) {
    const user = data.users[__VU % data.users.length];
    const headers = {
        'Content-Type': 'application/json',
    };

    let authStart = Date.now();
    let loginRes = http.post(`${BASE_URL}/auth`, JSON.stringify({
        username: user.username,
        password: user.password
    }), { headers });

    authDuration.add(Date.now() - authStart);

    check(loginRes, {
        'auth status is 200': (r) => r.status === 200,
        'has token': (r) => r.json('token') !== '',
    }) || errorRate.add(1);

    if (loginRes.status !== 200) {
        console.log(`Auth failed: ${loginRes.status} ${loginRes.body}`);
        return;
    }

    headers['Authorization'] = `Bearer ${loginRes.json('token')}`;

    let infoStart = Date.now();
    let infoRes = http.get(`${BASE_URL}/info`, { headers });

    infoDuration.add(Date.now() - infoStart);

    check(infoRes, {
        'info status is 200': (r) => r.status === 200,
        'has coins': (r) => r.json('coins') !== undefined,
    }) || errorRate.add(1);

    if (Math.random() < 0.3) {
        let sendStart = Date.now();
        let sendRes = http.post(`${BASE_URL}/sendCoin`, JSON.stringify({
            toUser: data.users[Math.floor(Math.random() * data.users.length)].username,
            amount: Math.floor(Math.random() * 100) + 1
        }), { headers });

        sendCoinDuration.add(Date.now() - sendStart);

        check(sendRes, {
            'send coins status is 200': (r) => r.status === 200,
        }) || errorRate.add(1);
    }

    if (Math.random() < 0.2) {
        let itemIndex = Math.floor(Math.random() * availableItems.length);
        let buyStart = Date.now();
        let buyRes = http.get(`${BASE_URL}/buy/${availableItems[itemIndex]}`, { headers });

        buyItemDuration.add(Date.now() - buyStart);

        check(buyRes, {
            'buy item status is 200': (r) => r.status === 200,
        }) || errorRate.add(1);
    }

    sleep(Math.random() * 0.1);
}
