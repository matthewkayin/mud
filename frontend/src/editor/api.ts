/* eslint-disable @typescript-eslint/no-explicit-any */

export async function apiWorldGetData(endpoint: string): Promise<any> {
  try {
    const response = await fetch(`http://${window.location.hostname}:7272${endpoint}`, {
    });
    if (!response.ok) {
      throw new Error(`HTTP response ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    return data;
  } catch (err) {
    console.error('Failed to fetch item data:', err);
    return;
  }
}
