// The view panel's standard line width, which is also the inventory table's width for the default capacity
export const VIEW_WIDTH = 70;

// A '+---- Title ----+' line, width columns wide, with the title centered
export const formatViewHeader = (title: string, width: number): string => {
  const paddedTitle = ` ${title} `;
  const dashCount = Math.max(width - 2 - paddedTitle.length, 0);
  const leftDashCount = Math.floor(dashCount / 2);
  const rightDashCount = dashCount - leftDashCount;
  return `+${'-'.repeat(leftDashCount)}${paddedTitle}${'-'.repeat(rightDashCount)}+`;
};
