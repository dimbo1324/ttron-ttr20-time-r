import { Activity, Gauge, History, Monitor, Settings, Zap } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { ComponentType } from 'react';
import { DashboardPage } from '../pages/dashboard/DashboardPage';
import { EmulatorPage } from '../pages/emulator/EmulatorPage';
import { GatewayPage } from '../pages/gateway/GatewayPage';
import { EventsPage } from '../pages/events/EventsPage';
import { SettingsPage } from '../pages/settings/SettingsPage';

export type RouteID = 'dashboard' | 'emulator' | 'gateway' | 'events' | 'settings';

export type Route = {
  id: RouteID;
  label: string;
  icon: LucideIcon;
  component: ComponentType;
};

export const routes: Route[] = [
  { id: 'dashboard', label: 'Dashboard', icon: Monitor, component: DashboardPage },
  { id: 'emulator', label: 'Emulator', icon: Zap, component: EmulatorPage },
  { id: 'gateway', label: 'Gateway', icon: Gauge, component: GatewayPage },
  { id: 'events', label: 'Events', icon: History, component: EventsPage },
  { id: 'settings', label: 'Settings', icon: Settings, component: SettingsPage },
];

export const appIcon = Activity;
