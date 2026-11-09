type InventoryProps = {
  isInWorld: boolean;
  lines: string[];
};

export const Inventory = ({ isInWorld, lines }: InventoryProps) => {
  return (
    <div className="inventory">
      {isInWorld && lines && lines.map((line) => (
        <p className="terminal-text">{line}</p>
      ))}
    </div>
  );
};
