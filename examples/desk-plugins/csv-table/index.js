// CSV artifact renderer: a Desk plugin in plain JS with no build step
// (#542). Desk sets `context` to { artifact, file } for each attachment
// whose type or extension matches plugin.json, and `api` to the plugin
// api (unused here). See docs/desk-plugins.md.

/** Splits CSV text into rows, honoring quoted fields and "" escapes. */
function parse(text) {
  const rows = [];
  let row = [], cell = '', quoted = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (quoted) {
      if (c === '"' && text[i + 1] === '"') { cell += '"'; i++; }
      else if (c === '"') quoted = false;
      else cell += c;
    } else if (c === '"') quoted = true;
    else if (c === ',') { row.push(cell); cell = ''; }
    else if (c === '\n' || c === '\r') {
      if (c === '\r' && text[i + 1] === '\n') i++;
      row.push(cell); rows.push(row); row = []; cell = '';
    } else cell += c;
  }
  if (cell !== '' || row.length) { row.push(cell); rows.push(row); }
  return rows;
}

const MAX_ROWS = 1000;

class CsvTable extends HTMLElement {
  set context(c) {
    this.load(c.file);
  }

  async load(file) {
    this.textContent = 'Loading ' + file.name + '…';
    try {
      const res = await fetch(file.url);
      if (!res.ok) throw new Error(res.status + ' ' + res.statusText);
      this.render(parse(await res.text()));
    } catch (e) {
      this.textContent = 'Could not load ' + file.name + ': ' + e.message;
    }
  }

  render(rows) {
    const table = document.createElement('table');
    table.className = 'csv-table';
    const [head = [], ...body] = rows;
    const tr = table.createTHead().insertRow();
    for (const h of head) {
      const th = document.createElement('th');
      th.textContent = h; // textContent, never innerHTML: the file is untrusted
      tr.append(th);
    }
    const tb = table.createTBody();
    for (const r of body.slice(0, MAX_ROWS)) {
      const row = tb.insertRow();
      for (const cell of r) row.insertCell().textContent = cell;
    }
    const note = document.createElement('div');
    note.className = 'faint';
    note.textContent = body.length > MAX_ROWS ? `first ${MAX_ROWS} of ${body.length} rows` : `${body.length} rows`;
    this.replaceChildren(table, note);
  }
}

customElements.define('csv-table-artifact', CsvTable);
