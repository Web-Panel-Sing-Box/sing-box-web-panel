export type ClientStatus = "active" | "disabled" | "expired";

export type Client = {
  id: string;
  name: string;
  uuid: string;
  inboundId: string;
  usedDown: number;
  usedUp: number;
  totalQuota: number;
  expiry: string;
  status: ClientStatus;
  subscription: string;
  startAfterFirstUse: boolean;
};

const GB = 1024 ** 3;

function uuid() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) return crypto.randomUUID();
  return "00000000-0000-4000-8000-000000000000";
}

function rndBytes(min: number, max: number) {
  return Math.floor(min + Math.random() * (max - min));
}

const NAMES = [
  "alex_kim",
  "kira",
  "jonas",
  "miyu",
  "lukas",
  "aria",
  "tomek",
  "noor",
  "yuki",
  "petra",
  "sasha",
  "louis",
  "leah",
  "rafael",
  "nora",
  "viktor",
  "ines",
  "dima",
  "anya",
  "mark",
  "rui",
  "tara",
  "elif",
  "ben"
];

const INBOUND_POOL = ["ib_01", "ib_02", "ib_03", "ib_04", "ib_05", "ib_06"];

function expiryFor(i: number) {
  const start = new Date("2026-06-01T00:00:00Z").getTime();
  const end = new Date("2026-12-31T00:00:00Z").getTime();
  const t = start + ((end - start) * ((i * 137) % 100)) / 100;
  return new Date(t).toISOString();
}

function quotaFor(i: number) {
  const tiers = [50 * GB, 100 * GB, 200 * GB, 500 * GB, 1024 * GB];
  return tiers[i % tiers.length];
}

function statusFor(i: number, used: number, quota: number, expiry: string): ClientStatus {
  if (new Date(expiry).getTime() < Date.now()) return "expired";
  if (i % 11 === 0) return "disabled";
  if (used / quota > 0.999) return "expired";
  return "active";
}

export function buildSeedClients(): Client[] {
  return NAMES.map((name, i) => {
    const quota = quotaFor(i);
    const used = rndBytes(Math.floor(quota * 0.05), Math.floor(quota * 0.97));
    const usedDown = Math.floor(used * 0.78);
    const usedUp = used - usedDown;
    const expiry = expiryFor(i);
    const id = `cl_${String(i + 1).padStart(2, "0")}`;
    return {
      id,
      name,
      uuid: uuid(),
      inboundId: INBOUND_POOL[i % INBOUND_POOL.length],
      usedDown,
      usedUp,
      totalQuota: quota,
      expiry,
      status: statusFor(i, used, quota, expiry),
      subscription: `https://panel.example/sub/${id}_${Math.random().toString(36).slice(2, 10)}`,
      startAfterFirstUse: i % 3 === 0
    };
  });
}
