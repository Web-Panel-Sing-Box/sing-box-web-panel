import { apiGet } from "./client";

export type SubscriptionLink = {
  label: string;
  url: string;
  protocol: string;
};

export type SubscriptionMeta = {
  name: string;
  used: number;
  total: number;
  expiry: string;
  online: boolean;
  subscriptionUrl: string;
  links: SubscriptionLink[];
};

// Public metadata for the human-facing subscription page. No auth required.
export function getSubscriptionMeta(token: string): Promise<SubscriptionMeta> {
  return apiGet<SubscriptionMeta>(`/subscription/${encodeURIComponent(token)}/meta`);
}
