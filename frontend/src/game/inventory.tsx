type InventoryProps = {
  isInWorld: boolean;
};

export const Inventory = ({ isInWorld }: InventoryProps) => {
  return (
    <div className="inventory">
      {isInWorld && <p className="terminal-text">Your inventory is empty.</p>}
    </div>
  );
};
