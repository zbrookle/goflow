import { DAG } from "../typing/dag_types"

const hostName = "http://localhost:8080";

function fetchJSON(url: string) {
  return fetch(url).then((res) => res.json());
}

export function fetchDAGs() {
  return fetchJSON(`${hostName}/dags`);
}

export function fetchDAG(dagName: string) {
  return fetchJSON(`${hostName}/dag/${dagName}`);
}

export function fetchDAGObject(dagName: string) {
  let dag = {} as DAG;
  fetchDAG(dagName).then((data) => {
    dag.config = data.Config;
    dag.isOn = data.isOn;
    dag.DAGRuns = data.DAGRuns;
  });
  return dag;
}
