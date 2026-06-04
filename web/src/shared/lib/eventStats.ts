import type { FrameEvent } from '../../entities/events/types';
export const eventDirections = ['RX', 'TX', 'ERR', 'SYSTEM'] as const;
export const eventSources = ['emulator', 'gateway'] as const;

export type EventStat = {
  label: string;
  count: number;
};

export function countEventsByDirection(events: FrameEvent[]): EventStat[] {
  return eventDirections.map((direction) => ({
    label: direction,
    count: events.filter((event) => event.direction === direction).length,
  }));
}

export function countEventsBySource(events: FrameEvent[], display: (source: string) => string): EventStat[] {
  return eventSources.map((source) => ({
    label: display(source),
    count: events.filter((event) => event.source === source || event.service === source).length,
  }));
}
