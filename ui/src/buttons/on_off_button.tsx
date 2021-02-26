import Switch from "bootstrap-switch-button-react";
import { useLayoutEffect, useState } from "react";

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
    const fetchURL = `http://localhost:8080/dag/${props.Name}`;
    fetch(fetchURL)
      .then((resp) => resp.json())
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
