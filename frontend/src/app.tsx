import { useTerminal } from './terminal'
import { useAuth } from './auth';

export const App = () => {
  const { containerRef, writeLine, setOnSubmit } = useTerminal('prompt>');
  const { token, loading, error, handleLogin } = useAuth();

  console.log('token ', token);

  setOnSubmit((command) => {
    writeLine(`You typed: ${command}`);
  });

  const TerminalDiv = () => (<div
    ref={containerRef}
    style={{
      width: '100vw',
      height: '600px',
      backgroundColor: '#1e1e1e',
      padding: '10px',
      borderRadius: '4px',
    }}
  />);

  return (
    <div>
      <h1>RC Disco MUD!</h1>
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
      }}>
        {loading && <p>Authorizing...</p>}
        {error && <p>Error: {error}</p>}
        {(token && token !== null) && <TerminalDiv />}
        {(!token && !loading) && <button onClick={handleLogin}>Login with Recurse</button>}
      </div>
    </div>
  )
};
