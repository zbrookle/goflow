import { useRef } from "react";

export const useComponentWillMount = (func: Function) => {
  const willMount = useRef(true);

  if (willMount.current) func();

  willMount.current = false;
};
