import { useState } from 'react';

type DebugLoginProps = {
  token: string | null;
  setToken: (value: string | null) => void;
}

export const DebugLogin = ({ token, setToken }: DebugLoginProps) => {
  const [tokenInput, setTokenInput] = useState('');

  const onButtonClick = () => {
    // The token will be a string, but validate that it is a proper number first
    const parsedToken = Number.parseInt(tokenInput);
    if (Number.isNaN(parsedToken)) {
      return;
    }
    if (parsedToken < 0) {
      return;
    }

    setToken(tokenInput);
  };

  return (
    <>
      {token === null &&
        <div>
          <input value={tokenInput} onChange={(event) => setTokenInput(event.target.value)}></input>
          <button onClick={onButtonClick}>Login</button>
        </div>}
    </>
  )
};
