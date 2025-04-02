import React, { useState } from "react";
import { v4 as uuid } from "uuid";
import { wrapper } from "./SayHello.module.scss";

type SayHelloProps = {

} & React.HTMLProps<HTMLDivElement>;

export const SayHello: React.FC<SayHelloProps> = () => {
  const [counter, setCounter] = useState(0);
  const incrementCounter = () => {
    setCounter((prev: number) => prev + 1);
  };
  const decrementCounter = () => {
    setCounter((prev: number) => prev - 1);
  };
  const resetCounter = () => {
    setCounter(0);
  };

  const randomString = uuid();

  return (
    <div className={wrapper}>
      <h1>Hello, world!</h1>
      <p>{randomString}</p>
      <p>Counter: {counter}</p>
      <button onClick={incrementCounter}>Increment</button>
      <button onClick={decrementCounter}>Decrement</button>
      <button onClick={resetCounter}>Reset</button>
    </div>
  );
}
