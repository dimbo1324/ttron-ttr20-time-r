import { Activity, BookOpen, Gauge, History, Monitor, Settings, Stethoscope, Zap } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { ComponentType } from 'react';
import type { TranslationKey } from '../shared/i18n/types';
import { DashboardPage } from '../pages/dashboard/DashboardPage';
import { EmulatorPage } from '../pages/emulator/EmulatorPage';
import { GatewayPage } from '../pages/gateway/GatewayPage';
import { EventsPage } from '../pages/events/EventsPage';
import { DiagnosticsPage } from '../pages/diagnostics/DiagnosticsPage';
import { GuidePage } from '../pages/guide/GuidePage';
import { SettingsPage } from '../pages/settings/SettingsPage';

export type RouteID = 'dashboard' | 'emulator' | 'gateway' | 'events' | 'diagnostics' | 'guide' | 'settings';

export type Route = {
  id: RouteID;
  labelKey: TranslationKey;
  icon: LucideIcon;
  component: ComponentType;
};

export const routes: Route[] = [
  { id: 'dashboard', labelKey: 'nav.dashboard', icon: Monitor, component: DashboardPage },
  { id: 'emulator', labelKey: 'nav.emulator', icon: Zap, component: EmulatorPage },
  { id: 'gateway', labelKey: 'nav.gateway', icon: Gauge, component: GatewayPage },
  { id: 'events', labelKey: 'nav.events', icon: History, component: EventsPage },
  { id: 'diagnostics', labelKey: 'nav.diagnostics', icon: Stethoscope, component: DiagnosticsPage },
  { id: 'guide', labelKey: 'nav.guide', icon: BookOpen, component: GuidePage },
  { id: 'settings', labelKey: 'nav.settings', icon: Settings, component: SettingsPage },
];

export const appIcon = Activity;
