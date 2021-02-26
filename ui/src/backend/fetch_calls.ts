const hostName = "http://localhost:8080"

function fetchJSON(url: string) {
    return fetch(url).then((res) => res.json())
}

export function fetchDAGs() {
    return fetchJSON(`${hostName}/dags`)
}

export function fetchDAG(dagName: string) {
    return fetchJSON(`${hostName}/dag/${dagName}`)
}