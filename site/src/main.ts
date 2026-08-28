import '@fontsource-variable/league-spartan';
import './styles.css';

const form = document.querySelector<HTMLFormElement>('#capsule-composer');
const output = document.querySelector<HTMLElement>('#review-output');
const copyButton = document.querySelector<HTMLButtonElement>('#copy-command');
const installInput = document.querySelector<HTMLInputElement>('#install-command');
const hostsInput = document.querySelector<HTMLInputElement>('#allowed-hosts');
const portsInput = document.querySelector<HTMLInputElement>('#allowed-ports');
const formError = document.querySelector<HTMLElement>('#form-error');

function values(): { install: string; hosts: string[]; ports: number[]; error: string } {
  const install = installInput?.value.trim() ?? '';
  const hosts = (hostsInput?.value ?? '').split(',').map((v) => v.trim().toLowerCase()).filter(Boolean);
  const rawPorts = (portsInput?.value ?? '').split(',').map((v) => v.trim()).filter(Boolean);
  const ports = rawPorts.map(Number);
  let error = '';
  if (!install) error = 'Add the command that retrieves or installs the project.';
  else if (hosts.some((host) => !/^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$/.test(host))) error = 'Use comma-separated hostnames without URLs or wildcards.';
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

form?.addEventListener('input', render);
form?.addEventListener('submit', (event) => event.preventDefault());
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
  connection.textContent = offline ? 'You’re offline. The guide remains available; downloads and remote installs need a connection.' : '';
}
window.addEventListener('online', renderConnection);
window.addEventListener('offline', renderConnection);
renderConnection();

if ('serviceWorker' in navigator && import.meta.env.PROD) {
  window.addEventListener('load', () => navigator.serviceWorker.register('/sw.js').catch(() => undefined));
}
