import type { DAGProps } from "../typing/dag_types"

type DagInfoProps = DAGProps & {
  Test: boolean
};

function DagInfo(props: DagInfoProps) {}

export { DagInfo };
