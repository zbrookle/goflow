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

type DAGConfig = {
  Name: string;
  Namespace: string;
  Schedule: string;
};

export type DAG = {
  config: DAGConfig;
  isOn: boolean;
};

export function fetchDAGObject(dagName: string) {
  let dag = {} as DAG;
  fetchDAG(dagName).then((data) => {
    dag.config = data.Config;
    dag.isOn = data.isOn;
  });
  return dag;
}
