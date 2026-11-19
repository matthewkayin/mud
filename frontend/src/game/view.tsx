import { useEffect, useRef } from 'react';
import { textLineToParagraph } from './text_line';
import { type ViewMode } from './view_mode';
import { formatInventoryTable, getInventoryTableWidth } from './inventory_table';
import { formatEquipmentTable } from './equipment_table';
import { VIEW_WIDTH, formatViewHeader } from './view_format';
import { type EquipmentSlot, type Inventory } from './socket_message';

const VIEW_TITLES: Record<ViewMode, string> = {
  help: 'Help',
  chat: 'Chat',
  inventory: 'Inventory',
};

const VIEW_HELP_LINES = [
  'This is the view panel. It can show one of several things at a time.',
  '\n',
  "Type 'view <mode>' to change what it shows. The modes are:",
  '  help      - this explanation',
  '  chat      - a history of chat messages',
  '  inventory - the items you have equipped and are carrying',
  "'view equipment' also shows the inventory.",
];

type ViewProps = {
  mode: ViewMode;
  isInWorld: boolean;
  chatLines: string[];
  inventory: Inventory | null;
  equipment: EquipmentSlot[] | null;
};

export const View = ({ mode, isInWorld, chatLines, inventory, equipment }: ViewProps) => {
  // A ref rather than state, since changing it never needs a re-render
  const shouldScrollToBottomRef = useRef(true);
  const containerRef = useRef<HTMLDivElement>(null);

  let lines: string[] = [];
  if (mode === 'help') {
    lines = VIEW_HELP_LINES;
  } else if (mode === 'chat') {
    lines = chatLines;
  } else if (mode === 'inventory' && inventory !== null && equipment !== null) {
    // The equipment table on top, stretched to the inventory table's width
    lines = [...formatEquipmentTable(equipment, getInventoryTableWidth(inventory)), ...formatInventoryTable(inventory)];
  }

  // Tables include their own header
  const isTable = mode === 'inventory';

  // Start the tables scrolled to the top, so the equipment stays visible, and every other mode scrolled to the bottom
  useEffect(() => {
    shouldScrollToBottomRef.current = !isTable;
    const container = containerRef.current;
    if (isTable && container !== null) {
      container.scrollTop = 0;
    }
  }, [mode, isTable]);

  // Outside the tables, keep following new lines while scrolled to the bottom
  useEffect(() => {
    const container = containerRef.current;
    if (container === null || isTable) {
      return;
    }

    // Set scrollTop directly rather than using scrollIntoView, which would also scroll the page
    if (shouldScrollToBottomRef.current) {
      container.scrollTop = container.scrollHeight;
    }
  }, [mode, isTable, isInWorld, chatLines, inventory, equipment]);

  const onScroll = (event: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = event.currentTarget;
    // scrollTop can be fractional on displays with a non-integer pixel ratio, so allow a small tolerance
    const distanceFromBottom = scrollHeight - clientHeight - scrollTop;
    shouldScrollToBottomRef.current = distanceFromBottom < 2;
  };

  return (
    <div className="view" ref={containerRef} onScroll={onScroll}>
      {isInWorld && !isTable && <div className="view-preformatted">{textLineToParagraph(formatViewHeader(VIEW_TITLES[mode], VIEW_WIDTH), 0)}</div>}
      {isInWorld && !isTable && lines.map(textLineToParagraph)}
      {isInWorld && isTable && <div className="view-preformatted">{lines.map(textLineToParagraph)}</div>}
    </div>
  );
};
