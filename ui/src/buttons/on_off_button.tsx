import Switch from "bootstrap-switch-button-react";
import { useLayoutEffect, useState } from "react";
import { fetchDAG } from "../backend/fetch_calls";

type OnOffButtonProps = {
  Name: string;
  IsOn: boolean;
};

export function OnOffButton(props: OnOffButtonProps) {
  const toggleRequestOptions = {
    method: "PUT",
  };
  const [isOn, setButtonOn] = useState(false);
  function getDAGIsOn() {
    fetchDAG(props.Name)
      .then((data) => setButtonOn(data.IsOn));
  }
  useLayoutEffect(getDAGIsOn, [props.Name]);
  return (
    <Switch
      size="sm"
      checked={isOn}
      onChange={() => {
        fetch(
          `http://localhost:8080/dag/${props.Name}/toggle`,
          toggleRequestOptions
        )
          .then((resp) => resp.json())
          .then((data) => setButtonOn(data.IsOn));
      }}
    />
  );
}
