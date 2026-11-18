import { useEffect, useRef } from 'react';
import { textLineToParagraph } from './text_line';
import { type ViewMode } from './view_mode';
import { formatInventoryTable } from './inventory_table';
import { type Inventory } from './socket_message';

const VIEW_TITLES: Record<ViewMode, string> = {
  help: 'Help',
  chat: 'Chat',
  inventory: 'Inventory',
  equipment: 'Equipment',
};

const VIEW_HELP_LINES = [
  'This is the view panel. It can show one of several things at a time.',
  '\n',
  "Type 'view <mode>' to change what it shows. The modes are:",
  '  help      - this explanation',
  '  chat      - a history of chat messages',
  '  inventory - the items you are carrying',
  '  equipment - the items you have equipped',
];

type ViewProps = {
  mode: ViewMode;
  isInWorld: boolean;
  chatLines: string[];
  inventory: Inventory | null;
  equipmentLines: string[] | null;
};

export const View = ({ mode, isInWorld, chatLines, inventory, equipmentLines }: ViewProps) => {
  // A ref rather than state, since changing it never needs a re-render
  const shouldScrollToBottomRef = useRef(true);
  const containerRef = useRef<HTMLDivElement>(null);

  let lines: string[] = [];
  if (mode === 'help') {
    lines = VIEW_HELP_LINES;
  } else if (mode === 'chat') {
    lines = chatLines;
  } else if (mode === 'inventory') {
    lines = inventory === null ? [] : formatInventoryTable(inventory);
  } else if (mode === 'equipment') {
    lines = equipmentLines ?? [];
  }

  // Tables have their own title and must not wrap
  const isTable = mode === 'inventory';

  // Start each mode scrolled to the bottom
  useEffect(() => {
    shouldScrollToBottomRef.current = true;
  }, [mode]);

  // Keep following new lines while scrolled to the bottom
  useEffect(() => {
    const container = containerRef.current;
    if (container === null) {
      return;
    }

    // Set scrollTop directly rather than using scrollIntoView, which would also scroll the page
    if (shouldScrollToBottomRef.current) {
      container.scrollTop = container.scrollHeight;
    }
  }, [mode, isInWorld, chatLines, inventory, equipmentLines]);

  const onScroll = (event: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = event.currentTarget;
    // scrollTop can be fractional on displays with a non-integer pixel ratio, so allow a small tolerance
    const distanceFromBottom = scrollHeight - clientHeight - scrollTop;
    shouldScrollToBottomRef.current = distanceFromBottom < 2;
  };

  return (
    <div className="view" ref={containerRef} onScroll={onScroll}>
      {isInWorld && !isTable && <p className="terminal-text">=== {VIEW_TITLES[mode]} ===</p>}
      {isInWorld && !isTable && lines.map(textLineToParagraph)}
      {isInWorld && isTable && <div className="view-table">{lines.map(textLineToParagraph)}</div>}
    </div>
  );
};
