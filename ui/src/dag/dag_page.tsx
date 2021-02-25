import type { DAGProps } from "../typing/dag_types";
import { Container, Row } from "react-bootstrap";
import { OnOffButton } from "../buttons/on_off_button";

type DagInfoProps = DAGProps & {
  JobName: string;
  MaxMemoryUsage: number;
  Successes: number;
  Failures: number;
};

function DagInfo(props: DagInfoProps) {
  return (
    <div>
      <OnOffButton Name={props.Name} IsOn={props.IsOn} />
      <h1 style={{ marginLeft: "1%" }}>{props.Name}</h1>
    </div>
  );
}

export { DagInfo };
