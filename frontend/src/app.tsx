import { useTerminal } from './terminal'

export const App = () => {
  const { containerRef, writeLine, setOnSubmit } = useTerminal('prompt> ');

  setOnSubmit((command) => {
    writeLine(`You typed: ${command}`);
  });

  return (
    <div>
      <h1>RC Disco MUD!</h1>
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
      }}>
        <div
          ref={containerRef}
          style={{
            width: '100vw',
            height: '600px',
            backgroundColor: '#1e1e1e',
            padding: '10px',
            borderRadius: '4px',
          }}
        />
      </div>
    </div>
  )
};
