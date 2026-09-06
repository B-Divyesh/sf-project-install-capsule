import '@fontsource-variable/league-spartan';
import './styles.css';

const form = document.querySelector<HTMLFormElement>('#capsule-composer');
const output = document.querySelector<HTMLElement>('#review-output');
const copyButton = document.querySelector<HTMLButtonElement>('#copy-command');
const installInput = document.querySelector<HTMLInputElement>('#install-command');
const hostsInput = document.querySelector<HTMLInputElement>('#allowed-hosts');
const portsInput = document.querySelector<HTMLInputElement>('#allowed-ports');
const formError = document.querySelector<HTMLElement>('#form-error');
const demoBanner = document.querySelector<HTMLElement>('#demo-banner');
const resetDemo = document.querySelector<HTMLButtonElement>('#reset-demo');
const startReal = document.querySelector<HTMLAnchorElement>('#start-real');
const routeAnnouncer = document.querySelector<HTMLElement>('#route-announcer');

const demoMode = new URLSearchParams(window.location.search).get('demo') === '1';
const demoStorageKey = 'demo:capsule-composer';
const sampleValues = {
  install: 'apk add --no-cache git python3 && git clone https://github.com/octocat/Hello-World.git .',
  hosts: 'codeload.github.com, dl-cdn.alpinelinux.org, github.com',
  ports: '3000'
};

function canonicalHost(value: string): string {
  return value.trim().toLowerCase().replace(/\.$/, '');
}

function isIPv4(value: string): boolean {
  const parts = value.split('.');
  return parts.length === 4 && parts.every((part) => /^\d{1,3}$/.test(part) && Number(part) <= 255);
}

function values(): { install: string; hosts: string[]; ports: number[]; error: string } {
  const install = installInput?.value.trim() ?? '';
  const hosts = (hostsInput?.value ?? '').split(',').map(canonicalHost).filter(Boolean);
  const rawPorts = (portsInput?.value ?? '').split(',').map((v) => v.trim()).filter(Boolean);
  const ports = rawPorts.map(Number);
  let error = '';
  if (!install) error = 'Add the command that retrieves or installs the project.';
  else if (hosts.some((host) => isIPv4(host) || !/^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$/.test(host))) error = 'Use lowercase hostnames without URLs, IP addresses, or wildcards.';
  else if (ports.some((port) => !Number.isInteger(port) || port < 1 || port > 65535)) error = 'Ports must be whole numbers from 1 to 65535.';
  return { install, hosts: [...new Set(hosts)], ports: [...new Set(ports)], error };
}

function shellQuote(value: string): string {
  return `'${value.replaceAll("'", `'"'"'`)}'`;
}

function render(): void {
  if (!form || !output) return;
  const state = values();
  if (formError) formError.textContent = state.error;
  form.toggleAttribute('data-invalid', Boolean(state.error));
  const hostArgs = state.hosts.map((host) => ` \\\n+  --allow-host ${host}`).join('');
  const portArgs = state.ports.map((port) => ` \\\n+  --port ${port}`).join('');
  const command = `capsule init \\\n+  --install ${shellQuote(state.install || 'git clone https://github.com/owner/project.git .')} \\\n+  --run 'npm install && npm run dev'${hostArgs}${portArgs}`;
  output.textContent = command;
  if (copyButton) copyButton.disabled = Boolean(state.error);
}

function persistDemo(): void {
  if (!demoMode || !installInput || !hostsInput || !portsInput) return;
  localStorage.setItem(demoStorageKey, JSON.stringify({ install: installInput.value, hosts: hostsInput.value, ports: portsInput.value }));
}

function loadSample(): void {
  if (!installInput || !hostsInput || !portsInput) return;
  installInput.value = sampleValues.install;
  hostsInput.value = sampleValues.hosts;
  portsInput.value = sampleValues.ports;
}

function enterDemo(): void {
  if (!demoMode) return;
  document.title = 'Demo — Project Install Capsule';
  demoBanner?.removeAttribute('hidden');
  form?.setAttribute('data-demo', 'true');
  let saved: Partial<typeof sampleValues> | undefined;
  try {
    saved = JSON.parse(localStorage.getItem(demoStorageKey) ?? 'null') ?? undefined;
  } catch {
    localStorage.removeItem(demoStorageKey);
  }
  loadSample();
  if (saved && installInput && hostsInput && portsInput) {
    installInput.value = typeof saved.install === 'string' ? saved.install : sampleValues.install;
    hostsInput.value = typeof saved.hosts === 'string' ? saved.hosts : sampleValues.hosts;
    portsInput.value = typeof saved.ports === 'string' ? saved.ports : sampleValues.ports;
  }
  routeAnnouncer && (routeAnnouncer.textContent = 'Sample capsule loaded.');
}

enterDemo();
form?.addEventListener('input', () => { render(); persistDemo(); });
form?.addEventListener('submit', (event) => event.preventDefault());
resetDemo?.addEventListener('click', () => {
  localStorage.removeItem(demoStorageKey);
  loadSample();
  render();
  routeAnnouncer && (routeAnnouncer.textContent = 'Sample capsule reset.');
  installInput?.focus();
});
startReal?.addEventListener('click', () => localStorage.removeItem(demoStorageKey));
copyButton?.addEventListener('click', async () => {
  if (!output) return;
  try {
    await navigator.clipboard.writeText(output.textContent ?? '');
    copyButton.textContent = 'Copied';
    window.setTimeout(() => { copyButton.textContent = 'Copy command'; }, 1800);
  } catch {
    copyButton.textContent = 'Select and copy';
    const selection = window.getSelection();
    const range = document.createRange();
    range.selectNodeContents(output);
    selection?.removeAllRanges();
    selection?.addRange(range);
  }
});
render();

const connection = document.querySelector<HTMLElement>('#connection-state');
function renderConnection(): void {
  if (!connection) return;
  const offline = !navigator.onLine;
  connection.hidden = !offline;
  connection.textContent = offline ? 'You’re offline. Remote install commands need a connection.' : '';
}
window.addEventListener('online', renderConnection);
window.addEventListener('offline', renderConnection);
renderConnection();

if ('serviceWorker' in navigator && import.meta.env.PROD) {
  window.addEventListener('load', () => navigator.serviceWorker.register('/sw.js').catch(() => undefined));
}
