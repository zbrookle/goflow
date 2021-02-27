export type DAGConfig = {
  Name: string;
  Namespace: string;
  Schedule: string;
  Command: Array<string>;
  Retries: number;
  DockerImage: string;
};

type DAGRun = {
  Name: string;
  StartTime: string;
  EndTime: string;
  ExecutionDate: string;
};

export type DAG = {
  config: DAGConfig;
  isOn: boolean;
  DAGRuns: Array<DAGRun>;
};

export type DAGProps = {
  dag: DAG;
};

export type DAGConfigProps = {
  config: DAGConfig;
};

type configKey = keyof DAGConfig
type configKeyArr = Array<configKey>

export function getConfigKeys(config: DAGConfig) {
  return Object.keys(config) as configKeyArr;
}
