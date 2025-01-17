import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
    discardResponseBodies: true,
    scenarios: {
        contacts: {
            executor: 'constant-vus',
            vus: 10,
            duration: '45s',
        },
    },
};

export default function () {
    const data = { game: 'Mobile Legends', gamerID: 'GYUTDTE', points: 20 }
    let res = http.post('http://localhost:8080/self', data)

    check(res, { 'ok': (r) => r.status === 200 })

    sleep(0.3)
}