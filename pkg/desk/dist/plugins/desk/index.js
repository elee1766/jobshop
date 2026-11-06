// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = "http://www.w3.org/2000/svg", i = Array.isArray, a = Array.prototype.indexOf, o = Array.prototype.includes, s = Array.from, c = Object.defineProperty, l = Object.getOwnPropertyDescriptor, u = Object.getOwnPropertyDescriptors, d = Object.prototype, f = Array.prototype, p = Object.getPrototypeOf, m = Object.isExtensible;
function h(e) {
	return typeof e == "function";
}
var g = () => {};
function _(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function v() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
function y(e, t, n = !1) {
	return e === void 0 ? n ? t() : t : e;
}
function b(e, t) {
	if (Array.isArray(e)) return e;
	if (t === void 0 || !(Symbol.iterator in e)) return Array.from(e);
	let n = [];
	for (let r of e) if (n.push(r), n.length === t) break;
	return n;
}
// node_modules/svelte/src/internal/client/constants.js
var x = 1 << 24, S = 1024, C = 2048, w = 4096, ee = 8192, te = 16384, ne = 32768, re = 1 << 25, ie = 65536, ae = 1 << 19, oe = 1 << 20, se = 1 << 25, ce = 1 << 21, le = 1 << 22, ue = 1 << 23, de = Symbol("$state"), fe = Symbol("component"), pe = Symbol("legacy props"), me = Symbol(""), he = Symbol("attributes"), ge = Symbol("class"), _e = Symbol("style"), ve = Symbol("text"), ye = Symbol("form reset"), be = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), xe = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function Se() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function Ce(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function we() {
	console.warn("https://svelte.dev/e/select_multiple_invalid_value");
}
function Te() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var T = !1;
function Ee(e) {
	T = e;
}
var E;
function D(t) {
	if (t === null) throw Ce(), e;
	return E = t;
}
function De() {
	return D(/* @__PURE__ */ un(E));
}
function O(t) {
	if (T) {
		if (/* @__PURE__ */ un(E) !== null) throw Ce(), e;
		E = t;
	}
}
function Oe(e = 1) {
	if (T) {
		for (var t = e, n = E; t--;) n = /* @__PURE__ */ un(n);
		E = n;
	}
}
function ke(e = !0) {
	for (var t = 0, n = E;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ un(n);
		e && n.remove(), n = i;
	}
}
function Ae(t) {
	if (!t || t.nodeType !== 8) throw Ce(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function je(e) {
	return e === this.v;
}
function Me(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Ne(e) {
	return !Me(e, this.v);
}
function Pe(e) {
	throw Error("https://svelte.dev/e/lifecycle_outside_component");
}
// node_modules/svelte/src/internal/client/errors.js
function Fe() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Ie(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Le(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Re() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function ze(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Be() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Ve(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function He() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ue() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function We() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Ge() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/shared/clone.js
var Ke = [];
function qe(e, t = !1, n = !1) {
	return Je(e, /* @__PURE__ */ new Map(), "", Ke, null, n);
}
function Je(e, t, n, r, a = null, o = !1) {
	if (typeof e == "object" && e) {
		var s = t.get(e);
		if (s !== void 0) return s;
		if (e instanceof Map) return new Map(e);
		if (e instanceof Set) return new Set(e);
		if (i(e)) {
			var c = Array(e.length);
			t.set(e, c), a !== null && t.set(a, c);
			for (var l = 0; l < e.length; l += 1) {
				var u = e[l];
				l in e && (c[l] = Je(u, t, n, r, null, o));
			}
			return c;
		}
		if (p(e) === d) {
			c = {}, t.set(e, c), a !== null && t.set(a, c);
			for (var f of Object.keys(e)) c[f] = Je(e[f], t, n, r, null, o);
			return c;
		}
		if (e instanceof Date) return e.getTime(), structuredClone(e);
		if (typeof e.toJSON == "function" && !o) return Je(e.toJSON(), t, n, r, e);
	}
	if (e instanceof EventTarget) return e;
	try {
		return structuredClone(e);
	} catch {
		return e;
	}
}
// node_modules/svelte/src/internal/shared/context.js
function Ye(e) {
	let t = e.p;
	for (; t !== null && t.c === null;) t = t.p;
	return t?.c ?? null;
}
function Xe(e, t) {
	return e === null && Pe(t), e.c ??= new Map(Ye(e) || void 0);
}
// node_modules/svelte/src/internal/client/context.js
var k = null;
function Ze(e) {
	k = e;
}
function Qe(e) {
	return Xe(k, "getContext").get(e);
}
function $e(e, t = !1, n) {
	k = {
		p: k,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: W,
		l: null
	};
}
function et(e) {
	var t = k, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) Tn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, k = t.p, tt(e);
}
function tt(e = {}) {
	return c(e, fe, { value: !0 }), e;
}
function nt() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var rt = [];
function it() {
	var e = rt;
	rt = [], _(e);
}
function at(e) {
	if (rt.length === 0 && !At) {
		var t = rt;
		queueMicrotask(() => {
			t === rt && it();
		});
	}
	rt.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var ot = ~(C | w | S);
function A(e, t) {
	e.f = e.f & ot | t;
}
function st(e) {
	e.f & 512 || e.deps === null ? A(e, S) : A(e, w);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function ct(e, t) {
	e.f & 2048 ? t.set(e, C) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, w), A(e, S);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function lt(e, t) {
	if (t) {
		let t = document.body;
		e.autofocus = !0, at(() => {
			document.activeElement === t && e.focus();
		});
	}
}
var ut = !1;
function dt() {
	ut || (ut = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[ye]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function ft(e) {
	var t = H, n = W;
	U(null), Kn(null);
	try {
		return e();
	} finally {
		U(t), Kn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function pt(e, t, n, r) {
	let i = nt() ? _t : bt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = W, c = mt(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				vn(e, s);
			}
			ht();
		}
	}
	var d = gt();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ yt(e))).then(u).catch((e) => vn(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), ht());
	}) : f();
}
function mt() {
	var e = W, t = H, n = k, r = M;
	return function(i = !0) {
		Kn(e), U(t), Ze(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function ht(e = !0) {
	Kn(null), U(null), Ze(null), e && M?.deactivate();
}
function gt() {
	var e = W, t = e.b, n = M, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function _t(e) {
	var n = 2 | C;
	return W !== null && (W.f |= ae), {
		ctx: k,
		deps: null,
		effects: null,
		equals: je,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: W,
		ac: null
	};
}
var vt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function yt(e, n, r) {
	let i = W;
	i === null && Fe();
	var a = void 0, o = Kt(t), s = !H, c = /* @__PURE__ */ new Set();
	return On(() => {
		var t = W, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== be && n.reject(e);
			}).finally(ht);
		} catch (e) {
			n.reject(e), ht();
		}
		var r = M;
		if (s) {
			if (t.f & 32768) var l = gt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(vt);
			else for (let e of c.values()) e.reject(vt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== vt && (r.activate(), t ? (o.f |= ue, Xt(o, t)) : (o.f & 8388608 && (o.f ^= ue), Xt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), Cn(() => {
		for (let e of c) e.reject(vt);
	}), new Promise((e) => {
		function t(n) {
			function r() {
				n === a ? e(o) : t(a);
			}
			n.then(r, r);
		}
		t(a);
	});
}
/*#__NO_SIDE_EFFECTS__*/
function j(e) {
	let t = /* @__PURE__ */ _t(e);
	return Jn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function bt(e) {
	let t = /* @__PURE__ */ _t(e);
	return t.equals = Ne, t;
}
function xt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) V(t[n]);
	}
}
function St(e) {
	var n, r = W, i = e.parent;
	if (!Un && i !== null && e.v !== t && i.f & 24576) return Se(), e.v;
	Kn(i);
	try {
		xt(e), n = rr(e);
	} finally {
		Kn(r);
	}
	return n;
}
function Ct(e) {
	var t = St(e);
	!e.equals(t) && (e.wv = er(), (!M?.is_fork || e.deps === null) && (M === null ? e.v = t : (M.capture(e, t, !0), Dt?.capture(e, t, !0)), e.deps === null)) ? A(e, S) : Un || (Ot === null ? st(e) : (Sn() || M?.is_fork) && Ot.set(e, t));
}
function wt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && ft(() => {
		t.ac.abort(be), t.ac = null;
	}), t.fn !== null && (t.teardown = g), or(t, 0), Nn(t));
}
function Tt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && sr(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var Et = null, M = null, Dt = null, Ot = null, kt = null, At = !1, jt = !1, Mt = null, Nt = null, Pt = 0, Ft = 1, It = class e {
	id = Ft++;
	started = !1;
	linked = !0;
	prev = null;
	next = null;
	async_deriveds = /* @__PURE__ */ new Map();
	current = /* @__PURE__ */ new Map();
	previous = /* @__PURE__ */ new Map();
	#e = /* @__PURE__ */ new Set();
	#t = /* @__PURE__ */ new Set();
	#n = 0;
	#r = /* @__PURE__ */ new Map();
	#i = null;
	#a = [];
	#o = [];
	#s = /* @__PURE__ */ new Map();
	#c = /* @__PURE__ */ new Map();
	#l = /* @__PURE__ */ new Set();
	is_fork = !1;
	#u = !1;
	constructor() {
		Et === null ? Et = this : (Et.next = this, this.prev = Et), Et = this;
	}
	#d() {
		if (this.is_fork) return !0;
		for (let e of this.#r.keys()) if (!this.#f(e)) return !0;
		return !1;
	}
	#f(e) {
		for (var t = e; t !== null; t = t.parent) if (this.#c.has(t)) return !0;
		return !1;
	}
	skip_effect(e) {
		this.#c.has(e) || this.#c.set(e, {
			d: [],
			m: []
		}), this.#l.delete(e);
	}
	unskip_effect(e, t = (e) => this.schedule(e)) {
		var n = this.#c.get(e);
		if (n) {
			this.#c.delete(e);
			for (var r of n.d) A(r, C), t(r);
			for (r of n.m) A(r, w), t(r);
		}
		this.#l.add(e);
	}
	#p() {
		var e = [];
		for (let i of this.#a) if (!(i.f & 16384 || !(i.f & 6144))) {
			for (var t = i, n = !1; t.parent !== null;) {
				t = t.parent;
				var r = t.f;
				if (r & 96) {
					if (!(r & 1024)) {
						n = !0;
						break;
					}
					t.f ^= S;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) A(e, t), this.schedule(e);
		this.apply();
		for (var t = Mt = [], n = [], r = Nt = []; this.#a.length > 0;) {
			Pt++ > 1e3 && (this.#b(), Lt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Ht(e), this.#d() || this.discard(), t;
			}
		}
		if (M = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (Mt = null, Nt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Vt(e, t);
			r.length > 0 && M.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), Dt = this, zt(n), zt(t), Dt = null, this.#i?.resolve();
			var o = M;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Wt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= S;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= S : i & 4 ? t.push(r) : tr(r) && (i & 16 && this.add_dirty_reaction(r, w), sr(r));
				var o = r.first;
				if (o !== null) {
					r = o;
					continue;
				}
			}
			for (; r !== null;) {
				var s = r.next;
				if (s !== null) {
					r = s;
					break;
				}
				r = r.parent;
			}
		}
	}
	#g() {
		for (var e = this.prev; e !== null; e = e.prev) if (!e.is_fork && this.#_(e)) return e;
		return null;
	}
	#_(e) {
		for (let [t, [, n]] of this.current) if (!n && e.current.has(t)) return !0;
		return !1;
	}
	#v(e) {
		for (let [t, n] of e.current) !this.previous.has(t) && e.previous.has(t) && this.previous.set(t, e.previous.get(t)), this.current.set(t, n);
		for (let [t, n] of e.async_deriveds) {
			let e = this.async_deriveds.get(t);
			e && n.promise.then(e.resolve).catch(e.reject);
		}
		e.async_deriveds.clear(), this.transfer_reactions(e.#s);
		let t = (e) => {
			var n = e.reactions;
			if (n !== null && !(e.f & 2 && !(e.f & 6144))) for (let e of n) {
				var r = e.f;
				if (r & 2) t(e);
				else {
					var i = e;
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), A(i, C), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), M = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) ct(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), Ot?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		M = this;
	}
	deactivate() {
		M = null, Ot = null;
	}
	flush() {
		try {
			jt = !0, M = this, this.#m();
		} finally {
			Pt = 0, kt = null, Mt = null, Nt = null, jt = !1, M = null, Ot = null, Wt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(vt);
		this.#b(), this.#i?.resolve();
	}
	register_created_effect(e) {
		this.#o.push(e);
	}
	increment(e, t) {
		if (this.#n += 1, e) {
			let e = this.#r.get(t) ?? 0;
			this.#r.set(t, e + 1);
		}
	}
	decrement(e, t) {
		if (--this.#n, e) {
			let e = this.#r.get(t) ?? 0;
			e === 1 ? this.#r.delete(t) : this.#r.set(t, e - 1);
		}
		this.#u || (this.#u = !0, at(() => {
			this.#u = !1, this.linked && this.flush();
		}));
	}
	transfer_reactions(e) {
		for (let [t, n] of e) this.add_dirty_reaction(t, n);
		e.clear();
	}
	add_dirty_reaction(e, t) {
		(t === 2048 || this.#s.get(e) !== 2048) && this.#s.set(e, t);
	}
	oncommit(e) {
		this.#e.add(e);
	}
	ondiscard(e) {
		this.#t.add(e);
	}
	settled() {
		return (this.#i ??= v()).promise;
	}
	static ensure() {
		if (M === null) {
			let t = M = new e();
			!jt && at(() => {
				t.started || t.flush();
			});
		}
		return M;
	}
	apply() {
		Ot = null;
	}
	schedule(e) {
		kt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? Et = e : t.prev = e, this.linked = !1;
		}
	}
};
function Lt() {
	try {
		Be();
	} catch (e) {
		vn(e, kt);
	}
}
var Rt = null;
function zt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && tr(r) && (Rt = /* @__PURE__ */ new Set(), sr(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && In(r), Rt?.size > 0)) {
				Wt.clear();
				for (let e of Rt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Rt.has(n) && (Rt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || sr(n);
					}
				}
				Rt.clear();
			}
		}
		Rt = null;
	}
}
function Bt(e) {
	M.schedule(e);
}
function Vt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), A(e, S);
		for (var n = e.first; n !== null;) Vt(n, t), n = n.next;
	}
}
function Ht(e) {
	A(e, S);
	for (var t = e.first; t !== null;) Ht(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Ut = /* @__PURE__ */ new Set(), Wt = /* @__PURE__ */ new Map(), Gt = !1;
function Kt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: je,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function N(e, t) {
	let n = Kt(e, t);
	return Jn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function qt(e, t = !1, n = !0) {
	let r = Kt(e);
	return t || (r.equals = Ne), r;
}
function P(e, t, n = !1) {
	return H !== null && (!Gn || H.f & 131072) && nt() && H.f & 4325394 && (qn === null || !qn.has(e)) && We(), Xt(e, n ? tn(t) : t, Nt);
}
var Jt = null, Yt = 0;
function Xt(e, t, n = null) {
	if (!e.equals(t)) {
		Un ? Wt.set(e, t) : Wt.has(e) || Wt.set(e, e.v);
		var r = It.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && St(t), Ot === null && st(t);
		}
		e.wv = er(), Jt = null, Yt = 0, en(e, C, n), Jt = null, nt() && W !== null && W.f & 1024 && !(W.f & 96) && (q === null ? Yn([e]) : q.push(e)), !r.is_fork && Ut.size > 0 && !Gt && Zt();
	}
	return t;
}
function Zt() {
	Gt = !1;
	for (let e of Ut) {
		e.f & 1024 && A(e, w);
		let t;
		try {
			t = tr(e);
		} catch {
			t = !0;
		}
		t && sr(e);
	}
	Ut.clear();
}
function Qt(e, t = 1) {
	var n = J(e), r = t === 1 ? n++ : n--;
	return P(e, n), r;
}
function $t(e) {
	P(e, e.v + 1);
}
function en(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = nt(), a = r.length;
		if (Yt += a, Yt > 1e5 && Jt === null && (Jt = /* @__PURE__ */ new Set()), Jt !== null) {
			if (Jt.has(e)) return;
			Jt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & C) === 0;
				if (l && A(s, t), c & 131072) Ut.add(s);
				else if (c & 2) {
					var u = s;
					Ot?.delete(u), en(u, w, n);
				} else if (l) {
					var d = s;
					c & 16 && Rt !== null && Rt.add(d), n === null ? Bt(d) : n.push(d);
				}
			}
		}
	}
}
function tn(e) {
	if (typeof e != "object" || !e || de in e || fe in e) return e;
	let n = p(e);
	if (n !== d && n !== f) return e;
	var r = /* @__PURE__ */ new Map(), a = i(e), o = /* @__PURE__ */ N(0), s = null, c = Qn, u = (e) => {
		if (Qn === c) return e();
		var t = H, n = Qn;
		U(null), $n(c);
		var r = e();
		return U(t), $n(n), r;
	};
	return a && r.set("length", /* @__PURE__ */ N(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || He();
			var i = r.get(t);
			return i === void 0 ? u(() => {
				var e = /* @__PURE__ */ N(n.value, s);
				return r.set(t, e), e;
			}) : P(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = u(() => /* @__PURE__ */ N(t, s));
					r.set(n, e), $t(o);
				}
			} else P(i, t), $t(o);
			return !0;
		},
		get(n, i, a) {
			if (i === de) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || l(n, i)?.writable) && (o = u(() => /* @__PURE__ */ N(tn(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = J(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = J(a);
				if (o === t) return;
				if (i && "value" in i) i.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return i;
		},
		has(e, n) {
			if (n === de) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (W !== null || H !== null) && (!a || l(e, n)?.writable)) && (i === void 0 && (i = u(() => /* @__PURE__ */ N(a ? tn(e[n]) : t, s)), r.set(n, i)), J(i) === t) ? !1 : a;
		},
		set(e, n, i, c) {
			var d = r.get(n), f = n in e;
			if (a && n === "length") for (var p = i; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = u(() => /* @__PURE__ */ N(t, s)), r.set(p + "", m)) : P(m, t);
			}
			if (d === void 0) (!f || l(e, n)?.writable) && (d = u(() => /* @__PURE__ */ N(void 0, s)), P(d, tn(i)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = u(() => tn(i));
				P(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, i), !f) {
				if (a && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && P(_, v + 1);
				}
				$t(o);
			}
			return !0;
		},
		ownKeys(e) {
			J(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = r.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [i, a] of r) a.v !== t && !(i in e) && n.push(i);
			return n;
		},
		setPrototypeOf() {
			Ue();
		}
	});
}
function nn(e) {
	try {
		if (typeof e == "object" && e && de in e) return e[de];
	} catch {}
	return e;
}
function rn(e, t) {
	return Object.is(nn(e), nn(t));
}
var an, on, sn, cn;
function ln() {
	if (an === void 0) {
		an = window, on = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		sn = l(t, "firstChild").get, cn = l(t, "nextSibling").get, m(e) && (e[ge] = void 0, e[he] = null, e[_e] = void 0, e.__e = void 0), m(n) && (n[ve] = void 0);
	}
}
function F(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function I(e) {
	return sn.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function un(e) {
	return cn.call(e);
}
function L(e, t) {
	if (!T) return /* @__PURE__ */ I(e);
	var n = /* @__PURE__ */ I(E);
	if (n === null) n = E.appendChild(F());
	else if (t && n.nodeType !== 3) {
		var r = F();
		return n?.before(r), D(r), r;
	}
	return t && gn(n), D(n), n;
}
function dn(e, t = !1) {
	if (!T) {
		var n = /* @__PURE__ */ I(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ un(n) : n;
	}
	if (t) {
		if (E?.nodeType !== 3) {
			var r = F();
			return E?.before(r), D(r), r;
		}
		gn(E);
	}
	return E;
}
function fn(e, t = !1) {
	if (!T) return /* @__PURE__ */ I(e);
	var n = L(e, t);
	return O(e), n;
}
function R(e, t = 1, n = !1) {
	let r = T ? E : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ un(r);
	if (!T) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = F();
			return r === null ? i?.after(a) : r.before(a), D(a), a;
		}
		gn(r);
	}
	return D(r), r;
}
function pn(e) {
	e.textContent = "";
}
function mn() {
	return !1;
}
function hn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function gn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function _n(e) {
	var t = W;
	if (t === null) return H.f |= ue, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	vn(e, t);
}
function vn(t, n) {
	if (t === e) throw t;
	if (!(n !== null && n.f & 16384)) {
		for (; n !== null;) {
			if (n.f & 128 && !(n.f & 33570816)) {
				if (!(n.f & 32768)) throw t;
				try {
					n.b.error(t);
					return;
				} catch (e) {
					t = e;
				}
			}
			n = n.parent;
		}
		throw t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/effects.js
function yn(e) {
	W === null && (H === null && ze(e), Re()), Un && Le(e);
}
function bn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function xn(e, t) {
	var n = W;
	n !== null && n.f & 8192 && (e |= ee);
	var r = {
		ctx: k,
		deps: null,
		nodes: null,
		f: e | C | 512,
		first: null,
		fn: t,
		last: null,
		next: null,
		parent: n,
		b: n && n.b,
		prev: null,
		teardown: null,
		wv: 0,
		ac: null
	};
	M?.register_created_effect(r);
	var i = r;
	if (e & 4) Mt === null ? It.ensure().schedule(r) : Mt.push(r);
	else if (t !== null) {
		try {
			sr(r);
		} catch (e) {
			throw V(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ie));
	}
	if (i !== null && (i.parent = n, n !== null && bn(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Sn() {
	return H !== null && !Gn;
}
function Cn(e) {
	let t = xn(8, null);
	return A(t, S), t.teardown = e, t;
}
function wn(e) {
	yn("$effect");
	var t = W.f;
	if (!H && t & 32 && k !== null && !k.i) {
		var n = k;
		(n.e ??= []).push(e);
	} else return Tn(e);
}
function Tn(e) {
	return xn(4 | oe, e);
}
function En(e) {
	It.ensure();
	let t = xn(64 | ae, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? Ln(t, () => {
			V(t), n(void 0);
		}) : (V(t), n(void 0));
	});
}
function Dn(e) {
	return xn(4, e);
}
function On(e) {
	return xn(le | ae, e);
}
function kn(e, t = 0) {
	return xn(8 | t, e);
}
function z(e, t = [], n = [], r = []) {
	pt(r, t, n, (t) => {
		xn(8, () => {
			e(...t.map(J));
		});
	});
}
function An(e, t = 0) {
	return xn(16 | t, e);
}
function jn(e, t = 0) {
	return xn(x | t, e);
}
function B(e) {
	return xn(32 | ae, e);
}
function Mn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Un, r = H;
		Wn(!0), U(null);
		try {
			t.call(null);
		} catch (t) {
			vn(t, e.parent);
		} finally {
			Wn(n), U(r);
		}
	}
}
function Nn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && ft(() => {
			e.abort(be);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : V(n, t), n = r;
	}
}
function Pn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || V(t), t = n;
	}
}
function V(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Fn(e.nodes.start, e.nodes.end), n = !0), e.f |= re, Nn(e, t && !n), or(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	Mn(e), e.f ^= re, e.f |= te;
	var i = e.parent;
	i !== null && i.first !== null && In(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Fn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ un(e);
		e.remove(), e = n;
	}
}
function In(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function Ln(e, t, n = !0) {
	var r = [];
	e.f |= 256, Rn(e, r, !0);
	var i = () => {
		n && V(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function Rn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ee;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				Rn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function zn(e) {
	e.f &= -257, Bn(e, !0);
}
function Bn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ee, e.f & 1024 || (A(e, C), It.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Bn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Vn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ un(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Hn = null, Un = !1;
function Wn(e) {
	Un = e;
}
var H = null, Gn = !1;
function U(e) {
	H = e;
}
var W = null;
function Kn(e) {
	W = e;
}
var qn = null;
function Jn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (qn ??= /* @__PURE__ */ new Set()).add(e);
}
var G = null, K = 0, q = null;
function Yn(e) {
	q = e;
}
var Xn = 1, Zn = 0, Qn = Zn;
function $n(e) {
	Qn = e;
}
function er() {
	return ++Xn;
}
function tr(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (tr(a) && Ct(a), a.wv > e.wv) return !0;
		}
		t & 512 && Ot === null && A(e, S);
	}
	return !1;
}
function nr(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(qn !== null && qn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? nr(a, t, !1) : t === a && (n ? A(a, C) : a.f & 1024 && A(a, w), Bt(a));
	}
}
function rr(e) {
	var t = G, n = K, r = q, i = H, a = qn, o = k, s = Gn, c = Qn, l = e.f;
	G = null, K = 0, q = null, H = l & 96 ? null : e, qn = null, Ze(e.ctx), Gn = !1, Qn = ++Zn, e.ac !== null && (ft(() => {
		e.ac.abort(be);
	}), e.ac = null);
	try {
		e.f |= ce;
		var u = e.fn, d = u();
		e.f |= ne;
		var f = ir(e);
		if (nt() && q !== null && !Gn && f !== null && !(e.f & 6146)) for (var p = 0; p < q.length; p++) nr(q[p], e);
		if (i !== null && i !== e) {
			if (Zn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Zn;
			if (t !== null) for (let e of t) e.rv = Zn;
			q !== null && (r === null ? r = q : r.push(...q));
		}
		return e.f & 8388608 && (e.f ^= ue), d;
	} catch (t) {
		return ir(e), _n(t);
	} finally {
		e.f ^= ce, G = t, K = n, q = r, H = i, qn = a, Ze(o), Gn = s, Qn = c;
	}
}
function ir(e) {
	var t = e.deps, n = M?.is_fork;
	if (G !== null) {
		var r;
		if (n || or(e, K), t !== null && K > 0) for (t.length = K + G.length, r = 0; r < G.length; r++) t[K + r] = G[r];
		else e.deps = t = G;
		if (Sn() && e.f & 512) for (r = K; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && K < t.length && (or(e, K), t.length = K);
	return t;
}
function ar(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = a.call(r, e);
		if (i !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[i] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (G === null || !o.call(G, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && st(c), c.ac !== null && ft(() => {
			c.ac.abort(be), c.ac = null, A(c, C);
		}), wt(c), or(c, 0);
	}
}
function or(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) ar(e, n[r]);
}
function sr(e) {
	var t = e.f;
	if (!(t & 16384)) {
		A(e, S);
		var n = W;
		W = e;
		try {
			t & 16777232 ? Pn(e) : Nn(e), Mn(e);
			var r = rr(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Xn;
		} finally {
			W = n;
		}
	}
}
function J(e) {
	var t = !!(e.f & 2);
	if (Hn?.add(e), H !== null && !Gn && !(W !== null && W.f & 16384) && (qn === null || !qn.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < Zn && (e.rv = Zn, G === null && n !== null && n[K] === e ? K++ : G === null ? G = [e] : G.push(e));
		else {
			H.deps ??= [], o.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : o.call(r, H) || r.push(H);
		}
	}
	if (Un && Wt.has(e)) return Wt.get(e);
	if (t) {
		var i = e;
		if (Un) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || lr(i)) && (a = St(i)), Wt.set(i, a), a;
		}
		var s = !(i.f & 512) && !Gn && H !== null && !!(H.f & 512), c = (i.f & ne) === 0;
		tr(i) && (s && (i.f |= 512), Ct(i)), s && !c && (Tt(i), cr(i));
	}
	if (Ot?.has(e)) return Ot.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function cr(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : o.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (Tt(n), cr(n));
	}
}
function lr(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Wt.has(t) || t.f & 2 && lr(t)) return !0;
	return !1;
}
function ur(e) {
	var t = Gn;
	try {
		return Gn = !0, e();
	} finally {
		Gn = t;
	}
}
function dr(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (de in e) fr(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && de in n && fr(n);
		}
	}
}
function fr(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			fr(e[n], t);
		} catch {}
		let n = p(e);
		if (n !== Object.prototype && n !== Array.prototype && n !== Map.prototype && n !== Set.prototype && n !== Date.prototype) {
			let t = u(n);
			for (let n in t) {
				let r = t[n].get;
				if (r) try {
					r.call(e);
				} catch {}
			}
		}
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var pr = Symbol("events"), mr = /* @__PURE__ */ new Set(), hr = /* @__PURE__ */ new Set();
function gr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || Sr.call(t, e), !e.cancelBubble) return ft(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, at(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function _r(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = gr(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && Cn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function vr(e, t, n) {
	(t[pr] ??= {})[e] = n;
}
function yr(e) {
	for (var t = 0; t < e.length; t++) mr.add(e[t]);
	for (var n of hr) n(e);
}
var br = null, xr = !1;
function Sr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	br = e, xr || (xr = !0, setTimeout(() => {
		xr = !1, br = null;
	}));
	var o = 0, s = br === e && e[pr];
	if (s) {
		var l = i.indexOf(s);
		if (l !== -1 && (t === document || t === window)) {
			e[pr] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (o = l);
	}
	if (a = i[o] || e.target, a !== t) {
		c(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = H, f = W;
		U(null), Kn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[pr]?.[r];
					h == null || a.disabled && e.target !== a || h.call(a, e);
				} catch (e) {
					p ? m.push(e) : p = e;
				}
				if (e.cancelBubble) break;
				o++, a = o < i.length ? i[o] : null;
			}
			if (p) {
				for (let e of m) queueMicrotask(() => {
					throw e;
				});
				throw p;
			}
		} finally {
			e[pr] = t, delete e.currentTarget, U(d), Kn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Cr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function wr(e) {
	return Cr?.createHTML(e) ?? e;
}
function Tr(e) {
	var t = hn("template");
	return t.innerHTML = wr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Er(e, t) {
	var n = W;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Y(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (T) return Er(E, null), E;
		i === void 0 && (i = Tr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ I(i)));
		var t = r || on ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ I(t), s = t.lastChild;
			Er(o, s);
		} else Er(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Dr(e, t, n = "svg") {
	var r = !e.startsWith("<!>"), i = !!(t & 1), a = `<${n}>${r ? e : "<!>" + e}</${n}>`, o;
	return () => {
		if (T) return Er(E, null), E;
		if (!o) {
			var e = /* @__PURE__ */ I(Tr(a));
			if (i) for (o = document.createDocumentFragment(); /* @__PURE__ */ I(e);) o.appendChild(/* @__PURE__ */ I(e));
			else o = /* @__PURE__ */ I(e);
		}
		var t = o.cloneNode(!0);
		if (i) {
			var n = /* @__PURE__ */ I(t), r = t.lastChild;
			Er(n, r);
		} else Er(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Or(e, t) {
	return /* @__PURE__ */ Dr(e, t, "svg");
}
function kr() {
	if (T) return Er(E, null), E;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = F();
	return e.append(t, n), Er(t, n), e;
}
function X(e, t) {
	if (T) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = E), De();
	} else e !== null && e.before(t);
}
// node_modules/svelte/src/utils.js
function Ar(e) {
	return e.endsWith("capture") && e !== "gotpointercapture" && e !== "lostpointercapture";
}
var jr = [
	"beforeinput",
	"click",
	"change",
	"dblclick",
	"contextmenu",
	"focusin",
	"focusout",
	"input",
	"keydown",
	"keyup",
	"mousedown",
	"mousemove",
	"mouseout",
	"mouseover",
	"mouseup",
	"pointerdown",
	"pointermove",
	"pointerout",
	"pointerover",
	"pointerup",
	"touchend",
	"touchmove",
	"touchstart"
];
function Mr(e) {
	return jr.includes(e);
}
var Nr = /* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split("."), Pr = {
	formnovalidate: "formNoValidate",
	ismap: "isMap",
	nomodule: "noModule",
	playsinline: "playsInline",
	readonly: "readOnly",
	defaultvalue: "defaultValue",
	defaultchecked: "defaultChecked",
	srcobject: "srcObject",
	novalidate: "noValidate",
	allowfullscreen: "allowFullscreen",
	disablepictureinpicture: "disablePictureInPicture",
	disableremoteplayback: "disableRemotePlayback"
};
function Fr(e) {
	return e = e.toLowerCase(), Pr[e] ?? e;
}
[...Nr];
var Ir = ["touchstart", "touchmove"];
function Lr(e) {
	return Ir.includes(e);
}
var Rr = [
	"textarea",
	"script",
	"style",
	"title"
];
function zr(e) {
	return Rr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Br(e) {
	let t = 0, n = Kt(0), r;
	return () => {
		Sn() && (J(n), kn(() => (t === 0 && (r = ur(() => e(() => $t(n)))), t += 1, () => {
			at(() => {
				--t, t === 0 && (r?.(), r = void 0, $t(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Vr = ie | ae;
function Hr(e, t, n, r) {
	new Ur(e, t, n, r);
}
var Ur = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = T ? E : null;
	#n;
	#r;
	#i;
	#a = null;
	#o = null;
	#s = null;
	#c = null;
	#l = 0;
	#u = 0;
	#d = !1;
	#f = /* @__PURE__ */ new Map();
	#p = null;
	#m = Br(() => (this.#p = Kt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = An(() => {
			if (T) {
				let e = Ae(this.#t);
				De();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Vr), T && (this.#e = E);
	}
	#h() {
		try {
			this.#a = B(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		at(r), t && (this.#s = B(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? Te() : (t = !0, n && Ge(), this.#s !== null && Ln(this.#s, () => {
				this.#s = null;
			}), this.#S(() => {
				this.#b();
			})));
		};
		return {
			reset: r,
			invoke_onerror: () => {
				if (!this.#v()) try {
					n = !0, this.#n.onerror?.(e, r), n = !1;
				} catch (e) {
					vn(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (te | re)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = B(() => e(this.#e)), at(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = F(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return B(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							vn(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(M)) : this.#u === 0 && (this.#e.before(e), this.#c = null, Ln(this.#o, () => {
					this.#o = null;
				}), this.#x(M));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = B(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Vn(this.#a, e);
				let t = this.#n.pending;
				this.#o = B(() => t(this.#e));
			} else this.#x(M);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		ct(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = W, n = H, r = k;
		Kn(this.#i), U(this.#i), Ze(this.#i.ctx);
		try {
			return It.ensure(), e();
		} finally {
			Kn(t), U(n), Ze(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && Ln(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, at(() => {
			this.#d = !1, this.#p && Xt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), J(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		M?.is_fork ? (this.#a && M.skip_effect(this.#a), this.#o && M.skip_effect(this.#o), this.#s && M.skip_effect(this.#s), M.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (V(this.#a), null), this.#o &&= (V(this.#o), null), this.#s &&= (V(this.#s), null), T && (D(this.#t), Oe(), D(ke()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return B(() => {
						var r = W;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return vn(e, this.#i.parent), null;
				}
			}));
		};
		at(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					vn(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => vn(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Wr(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[ve] ??= e.nodeValue) && (e[ve] = n, e.nodeValue = `${n}`);
}
function Gr(e, t) {
	return qr(e, t);
}
var Kr = /* @__PURE__ */ new Map();
function qr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: c = !0, transformError: l }) {
	ln();
	var u = void 0, d = En(() => {
		var c = r ?? n.appendChild(F());
		Hr(c, { pending: () => {} }, (n) => {
			$e({});
			var r = k;
			if (o && (r.c = o), a && (i.$$events = a), T && Er(n, null), u = t(n, i) || tt(), T && (W.nodes.end = E, E === null || E.nodeType !== 8 || E.data !== "]")) throw Ce(), e;
			et();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Lr(r);
					for (let e of [n, document]) {
						var a = Kr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Kr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Sr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(s(mr)), hr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Kr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Sr), t.delete(e), t.size === 0 && Kr.delete(r)) : t.set(e, i);
			}
			hr.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return Jr.set(u, d), u;
}
var Jr = /* @__PURE__ */ new WeakMap();
function Yr(e, t) {
	let n = Jr.get(e);
	return n ? (Jr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Xr = class {
	anchor;
	#e = /* @__PURE__ */ new Map();
	#t = /* @__PURE__ */ new Map();
	#n = /* @__PURE__ */ new Map();
	#r = /* @__PURE__ */ new Set();
	#i = !0;
	constructor(e, t = !0) {
		this.anchor = e, this.#i = t;
	}
	#a = (e) => {
		if (this.#e.has(e)) {
			var t = this.#e.get(e), n = this.#t.get(t);
			if (n) zn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (zn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (V(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Vn(r, t), t.append(F()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else V(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), Ln(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (V(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = M, r = mn();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = F();
				i.append(a), this.#n.set(e, {
					effect: B(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, B(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else T && (this.anchor = E), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/snippet.js
function Zr(e, t, ...n) {
	var r = new Xr(e);
	An(() => {
		let e = t() ?? null;
		r.ensure(e, e && ((t) => e(t, ...n)));
	}, ie);
}
function Qr(e) {
	k === null && Pe("onMount"), wn(() => {
		let t = ur(e);
		if (typeof t == "function") return t;
	});
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function Z(e, t, n = !1) {
	var r;
	T && (r = E, De());
	var i = new Xr(e), a = n ? ie : 0;
	function o(e, t) {
		if (T) {
			var n = Ae(r);
			if (e !== parseInt(n.substring(1))) {
				var a = ke();
				D(a), i.anchor = a, Ee(!1), i.ensure(e, t), Ee(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	An(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/key.js
var $r = Symbol("NaN");
function ei(e, t, n) {
	T && De();
	var r = new Xr(e), i = !nt();
	An(() => {
		var e = t();
		e !== e && (e = $r), i && typeof e == "object" && e && (e = {}), r.ensure(e, n);
	});
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function ti(e, t) {
	return t;
}
function ni(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, c = 0; c < i; c++) {
		let n = t[c];
		Ln(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					ri(e, s(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			pn(d), d.append(u), e.items.clear();
		}
		ri(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function ri(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= se, Vn(a, document.createDocumentFragment())) : V(t[i], n);
	}
}
var ii;
function ai(e, t, n, r, a, o = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = T ? D(/* @__PURE__ */ I(u)) : u.appendChild(F());
	}
	T && De();
	var d = null, f = /* @__PURE__ */ bt(() => {
		var e = n();
		return i(e) ? e : e == null ? [] : s(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = J(f);
			_.fallback = d, si(_, n, c, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= se, li(d, null, c)) : zn(d) : Ln(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: An(() => {
			var e = J(f), i = e.length;
			let s = !1;
			T && Ae(c) === "[!" != (i === 0) && (c = ke(), D(c), Ee(!1), s = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = M, v = mn(), y = 0; y < i; y += 1) {
				T && E.nodeType === 8 && E.data === "]" && (c = E, s = !0, Ee(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Xt(S.v, b), S.i && Xt(S.i, y), v && _.unskip_effect(S.e)) : (S = ci(l, m ? c : ii ??= F(), b, x, y, a, t, n), m || (S.e.f |= se), l.set(x, S)), u.add(x);
			}
			if (i === 0 && o && !d && (m ? d = B(() => o(c)) : (d = B(() => o(ii ??= F())), d.f |= se)), i > u.size && Ie("", "", ""), T && i > 0 && D(ke()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			s && Ee(!0), J(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, T && (c = E);
}
function oi(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function si(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, c = e.items, l = oi(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (zn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= se, _ === l) li(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), ui(e, d, _), ui(e, _, y), li(_, y, n), d = _, p = [], m = [], l = oi(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) li(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					ui(e, S.prev, C.next), ui(e, d, S), ui(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), li(_, l, n), ui(e, _.prev, _.next), ui(e, _, d === null ? e.effect.first : d.next), ui(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = oi(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = oi(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (ri(e, s(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = oi(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			ni(e, w, te);
		}
	}
	a && at(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function ci(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Kt(n) : /* @__PURE__ */ qt(n, !1, !1) : null, l = o & 2 ? Kt(i) : null;
	return {
		v: c,
		i: l,
		e: B(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function li(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ un(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function ui(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-component.js
function di(e, t, n) {
	var r;
	T && (r = E, De());
	var i = new Xr(e);
	An(() => {
		var e = t() ?? null;
		if (T && Ae(r) === "[" != (e !== null)) {
			var a = ke();
			D(a), i.anchor = a, Ee(!1), i.ensure(e, e && ((t) => n(t, e))), Ee(!0);
		} else i.ensure(e, e && ((t) => n(t, e)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-element.js
function fi(e, t, n, i, a, o) {
	let s = T;
	T && De();
	var c = null;
	T && E.nodeType === 1 && (c = E, De());
	var l = T ? E : e, u = new Xr(l, !1);
	An(() => {
		let e = t() || null;
		var o = a ? a() : n || e === "svg" ? r : void 0;
		if (e === null) u.ensure(null, null);
		else return u.ensure(e, (t) => {
			if (e) {
				var n = T;
				if (c = T ? c : hn(e, o), Er(c, c), i) {
					var r = null;
					T && zr(e) && c.append(r = document.createComment(""));
					var a = T ? /* @__PURE__ */ I(c) : c.appendChild(F());
					T && (a === null ? Ee(!1) : D(a)), i(c, a), r?.remove();
				}
				W.nodes.end = c, n || t.before(c);
			}
			T && D(t);
		}), () => {};
	}, ie), Cn(() => {}), s && (Ee(!0), D(l));
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function pi(e, t, n) {
	Dn(() => {
		var r = ur(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			kn(() => {
				var e = n();
				dr(e), i && Me(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/client/dom/elements/attachments.js
function mi(e, t) {
	var n = void 0, r;
	jn(() => {
		n !== (n = t()) && (r &&= (V(r), null), n && (r = B(() => {
			Dn(() => n(e));
		})));
	});
}
// node_modules/clsx/dist/clsx.mjs
function hi(e) {
	var t, n, r = "";
	if (typeof e == "string" || typeof e == "number") r += e;
	else if (typeof e == "object") {
		if (Array.isArray(e)) {
			var i = e.length;
			for (t = 0; t < i; t++) e[t] && (n = hi(e[t])) && (r && (r += " "), r += n);
		} else for (n in e) e[n] && (r && (r += " "), r += n);
	}
	return r;
}
function gi() {
	for (var e, t, n = 0, r = "", i = arguments.length; n < i; n++) (e = arguments[n]) && (t = hi(e)) && (r && (r += " "), r += t);
	return r;
}
// node_modules/svelte/src/internal/shared/attributes.js
function _i(e) {
	return typeof e == "object" ? gi(e) : e ?? "";
}
var vi = [..." 	\n\r\f\xA0\v﻿"];
function yi(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || vi.includes(r[o - 1])) && (s === r.length || vi.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
function bi(e, t = !1) {
	var n = t ? " !important;" : ";", r = "";
	for (var i of Object.keys(e)) {
		var a = e[i];
		a != null && a !== "" && (r += " " + i + ": " + a + n);
	}
	return r;
}
function xi(e) {
	return e[0] !== "-" || e[1] !== "-" ? e.toLowerCase() : e;
}
function Si(e, t) {
	if (t) {
		var n = "", r, i;
		if (Array.isArray(t) ? (r = t[0], i = t[1]) : r = t, e) {
			e = String(e).replaceAll(/\/\*.*?\*\//g, "").trim();
			var a = !1, o = 0, s = !1, c = [];
			r && c.push(...Object.keys(r).map(xi)), i && c.push(...Object.keys(i).map(xi));
			var l = 0, u = -1;
			let t = e.length;
			for (var d = 0; d < t; d++) {
				var f = e[d];
				if (s ? f === "/" && e[d - 1] === "*" && (s = !1) : a ? a === f && (a = !1) : f === "/" && e[d + 1] === "*" ? s = !0 : f === "\"" || f === "'" ? a = f : f === "(" ? o++ : f === ")" && o--, !s && a === !1 && o === 0) {
					if (f === ":" && u === -1) u = d;
					else if (f === ";" || d === t - 1) {
						if (u !== -1) {
							var p = xi(e.substring(l, u).trim());
							if (!c.includes(p)) {
								f !== ";" && d++;
								var m = e.substring(l, d).trim();
								n += " " + m + ";";
							}
						}
						l = d + 1, u = -1;
					}
				}
			}
		}
		return r && (n += bi(r)), i && (n += bi(i, !0)), n = n.trim(), n === "" ? null : n;
	}
	return e == null ? null : String(e);
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Ci(e, t, n, r, i, a) {
	var o = e[ge];
	if (T || o !== n || o === void 0) {
		var s = yi(n, r, a);
		T && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ge] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/style.js
function wi(e, t = {}, n, r) {
	for (var i in n) {
		var a = n[i];
		t[i] !== a && (n[i] == null ? e.style.removeProperty(i) : e.style.setProperty(i, a, r));
	}
}
function Ti(e, t, n, r) {
	var i = e[_e];
	if (T || i !== t) {
		var a = Si(t, r);
		T && a === e.getAttribute("style") || (a == null ? e.removeAttribute("style") : e.style.cssText = a), e[_e] = t;
	} else r && (Array.isArray(r) ? (wi(e, n?.[0], r[0]), wi(e, n?.[1], r[1], "important")) : wi(e, n, r));
	return r;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function Ei(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function Di(e, t) {
	var n = !("__defaultValue" in e);
	(n || e.__defaultValue !== t) && (e.__defaultValue = t, Oi(e, !n || "__value" in e));
}
function Oi(e, t) {
	var n = e.__defaultValue, r = e.multiple, a = r ? n ?? [] : null;
	if (!r || i(a)) {
		var o = e.selectedIndex, s = t && r ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = ji(c);
			Ei(c, r ? a.includes(l) : rn(l, n));
		}
		if (t) {
			if (s !== null) for (c of e.options) {
				var u = s.has(c);
				c.selected !== u && (c.selected = u);
			}
			else e.selectedIndex !== o && (e.selectedIndex = o);
		}
	}
}
function ki(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!i(t)) return we();
		for (var r of e.options) r.selected = t.includes(ji(r));
	} else {
		for (r of e.options) if (rn(ji(r), t)) {
			r.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function Ai(e) {
	var t = new MutationObserver((t) => {
		t.every(Mi) || ("__defaultValue" in e && Oi(e, !1), "__value" in e && ki(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), Cn(() => {
		t.disconnect();
	});
}
function ji(e) {
	return "__value" in e ? e.__value : e.value;
}
function Mi(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Ni = Symbol("class"), Pi = Symbol("style"), Fi = Symbol("is custom element"), Ii = Symbol("is html"), Li = xe ? "link" : "LINK", Ri = xe ? "input" : "INPUT", zi = xe ? "option" : "OPTION", Bi = xe ? "select" : "SELECT", Vi = xe ? "progress" : "PROGRESS";
function Hi(e) {
	if (T) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Q(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Q(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ye] = n, at(n), dt();
	}
}
function Q(e, t, n, r) {
	var i = Gi(e);
	if (!(T && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Li))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[me] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && qi(e).has(t) ? e[t] = n : (!T || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Ui(e, n, r, i, a = !1, o = !1) {
	T && a && e.nodeName === Ri && ("defaultValue" in r || "defaultChecked" in r || Hi(e));
	var s = Gi(e), c = s[Fi], l = !s[Ii];
	let u = T && c;
	u && Ee(!1);
	var d = n || {}, f = e.nodeName === zi, p = e.nodeName === Bi;
	try {
		for (var m in n) !(m in r) && m[0] + m[1] !== "$$" && (r[m] = null);
		r.class ? r.class = _i(r.class) : (i || r[Ni]) && (r.class = null), r[Pi] && (r.style ??= null);
		var h = qi(e);
		if (e.nodeName === Ri && "type" in r && ("value" in r || "__value" in r)) {
			var g = r.type;
			(g !== d.type || g === void 0 && e.hasAttribute("type")) && (d.type = g, Q(e, "type", g, o));
		}
		for (let a in r) {
			let u = r[a];
			if (f && a === "value" && u == null) e.value = e.__value = "", d[a] = u;
			else if (a === "class") Ci(e, e.namespaceURI === "http://www.w3.org/1999/xhtml", u, i, n?.[Ni], r[Ni]), d[a] = u, d[Ni] = r[Ni];
			else if (a === "style") Ti(e, u, n?.[Pi], r[Pi]), d[a] = u, d[Pi] = r[Pi];
			else {
				var _ = d[a];
				if (u !== _ || u === void 0 && e.hasAttribute(a)) {
					d[a] = u;
					var v = a[0] + a[1];
					if (v !== "$$") {
						if (v === "on") {
							let t = {}, n = "$$" + a, r = a.slice(2);
							var y = Mr(r);
							if (Ar(r) && (r = r.slice(0, -7), t.capture = !0), !y && _) {
								if (u != null) continue;
								e.removeEventListener(r, d[n], t), d[n] = null;
							}
							if (y) vr(r, e, u), yr([r]);
							else if (u != null) {
								function i(e) {
									d[a].call(this, e);
								}
								d[n] = gr(r, e, i, t);
							}
						} else if (a === "style") Q(e, a, u);
						else if (a === "autofocus") lt(e, !!u);
						else if (!c && (a === "__value" || a === "value" && u != null)) e.__value = u, (_ == null || e.value !== u || u === 0 && e.nodeName === Vi) && (e.value = u);
						else if (a === "selected" && f) Ei(e, u);
						else {
							var b = a;
							l || (b = Fr(b));
							var x = b === "defaultValue" || b === "defaultChecked";
							if (p && b === "defaultValue") continue;
							if (u == null && !c && !x) {
								if (s[a] = null, b === "value" || b === "checked") {
									let t = e, r = n === void 0;
									if (b === "value") {
										let e = t.defaultValue;
										t.removeAttribute(b), t.defaultValue = e, t.value = t.__value = r ? e : null;
									} else {
										let e = t.defaultChecked;
										t.removeAttribute(b), t.defaultChecked = e, t.checked = r ? e : !1;
									}
								} else e.removeAttribute(a);
							} else x || (c || typeof u != "string") && h.has(b) ? (e[b] = u, b in s && (s[b] = t)) : typeof u != "function" && Q(e, b, u, o);
						}
					}
				}
			}
		}
	} finally {
		u && Ee(!0);
	}
	return d;
}
function Wi(e, t, n = [], r = [], i = [], a, o = !1, s = !1) {
	pt(i, n, r, (n) => {
		var r = void 0, i = {}, c = e.nodeName === Bi, l = !1;
		if (jn(() => {
			var u = t(...n.map(J)), d = Ui(e, r, u, a, o, s);
			if (l && c) {
				var f = e;
				"defaultValue" in u && Di(f, u.defaultValue), "value" in u && ki(f, u.value);
			}
			for (let e of Object.getOwnPropertySymbols(i)) u[e] || V(i[e]);
			for (let t of Object.getOwnPropertySymbols(u)) {
				var p = u[t];
				t.description !== "@attach" || r && p === r[t] || (i[t] && V(i[t]), i[t] = B(() => mi(e, () => p))), d[t] = p;
			}
			r = d;
		}), c) {
			var u = e;
			Dn(() => {
				var e = r;
				"defaultValue" in e && Di(u, e.defaultValue), ki(u, e.value, !0), Ai(u);
			});
		}
		l = !0;
	});
}
function Gi(e) {
	return e[he] ??= {
		[Fi]: e.nodeName.includes("-"),
		[Ii]: e.namespaceURI === n
	};
}
var Ki = /* @__PURE__ */ new Map();
function qi(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Ki.get(t);
	if (n) return n;
	Ki.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = u(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = p(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/this.js
function Ji(e, t) {
	return e === t || e?.[de] === t;
}
function Yi(e = tt(), t, n, r) {
	var i = k.r, a = W;
	return Dn(() => {
		var o, s;
		return kn(() => {
			o = s, s = r?.() || [], ur(() => {
				Ji(n(...s), e) || (t(e, ...s), o && Ji(n(...o), e) && t(null, ...o));
			});
		}), () => {
			let r = a;
			for (; r !== i && r.parent !== null && r.parent.f & 33554432;) r = r.parent;
			let o = () => {
				s && Ji(n(...s), e) && t(null, ...s);
			}, c = r.teardown;
			r.teardown = () => {
				o(), c?.();
			};
		};
	}), e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Xi = !1;
function Zi(e) {
	var t = Xi;
	try {
		return Xi = !1, [e(), Xi];
	} finally {
		Xi = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
var Qi = {
	get(e, t) {
		if (!e.exclude.has(t)) return e.props[t];
	},
	set(e, t) {
		return !1;
	},
	getOwnPropertyDescriptor(e, t) {
		if (!e.exclude.has(t) && t in e.props) return {
			enumerable: !0,
			configurable: !0,
			value: e.props[t]
		};
	},
	has(e, t) {
		return !e.exclude.has(t) && t in e.props;
	},
	ownKeys(e) {
		return Reflect.ownKeys(e.props).filter((t) => !e.exclude.has(t));
	}
};
/*#__NO_SIDE_EFFECTS__*/
function $(e, t, n) {
	return new Proxy({
		props: e,
		exclude: t
	}, Qi);
}
var $i = {
	get(e, t) {
		let n = e.props.length;
		for (; n--;) {
			let r = e.props[n];
			if (h(r) && (r = r()), typeof r == "object" && r && t in r) return r[t];
		}
	},
	set(e, t, n) {
		let r = e.props.length;
		for (; r--;) {
			let i = e.props[r];
			h(i) && (i = i());
			let a = l(i, t);
			if (a && a.set) return a.set(n), !0;
		}
		return !1;
	},
	getOwnPropertyDescriptor(e, t) {
		let n = e.props.length;
		for (; n--;) {
			let r = e.props[n];
			if (h(r) && (r = r()), typeof r == "object" && r && t in r) {
				let e = l(r, t);
				return e && !e.configurable && (e.configurable = !0), e;
			}
		}
	},
	has(e, t) {
		if (t === de || t === pe) return !1;
		for (let n of e.props) if (h(n) && (n = n()), n != null && t in n) return !0;
		return !1;
	},
	ownKeys(e) {
		let t = [];
		for (let n of e.props) if (h(n) && (n = n()), n) {
			for (let e in n) t.includes(e) || t.push(e);
			for (let e of Object.getOwnPropertySymbols(n)) t.includes(e) || t.push(e);
		}
		return t;
	}
};
function ea(...e) {
	return new Proxy({ props: e }, $i);
}
function ta(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ _t(r), J(u)) : (c && (c = !1, s = o ? ur(r) : r), s);
	let f;
	if (a) {
		var p = de in e || pe in e;
		f = l(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Zi(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ve(t), f(m)));
	var g = i ? () => {
		var n = e[t];
		return n === void 0 ? d() : (c = !0, n);
	} : () => {
		var n = e[t];
		return n !== void 0 && (s = void 0), n === void 0 ? s : n;
	};
	if (i && !(n & 4)) return g;
	if (f) {
		var _ = e.$$legacy;
		return (function(e, t) {
			return arguments.length > 0 ? ((!i || !t || _ || h) && f(t ? g() : e), e) : g();
		});
	}
	var v = !1, y = (n & 1 ? _t : bt)(() => (v = !1, g()));
	a && J(y);
	var b = W;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? J(y) : i && a ? tn(e) : e;
			return P(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Un && v || b.f & 16384 ? y.v : J(y);
	});
}
// kit/plugin.svelte.js
function na(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href, r = null, i = !1, a = () => (r ??= new Promise((e) => {
		let t = document.querySelector(`link[href="${n}"]`) ?? document.createElement("link");
		t.sheet ? e() : (t.addEventListener("load", () => e(), { once: !0 }), t.addEventListener("error", () => e(), { once: !0 }), t.isConnected || (t.rel = "stylesheet", t.href = n, document.head.append(t)));
	}).then(() => {
		i = !0;
	}), r);
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ N(tn({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return J(this.#e);
		}
		set props(e) {
			P(this.#e, e, !0);
		}
		app = null;
		set api(e) {
			this.props.api = e;
		}
		get api() {
			return this.props.api;
		}
		set context(e) {
			this.props.context = e ?? {};
		}
		get context() {
			return this.props.context;
		}
		set config(e) {
			this.props.config = e ?? {};
		}
		get config() {
			return this.props.config;
		}
		connectedCallback() {
			this.style.display = "contents";
			let e = a();
			this.app ||= Gr(t, {
				target: this,
				props: this.props
			}), i || (this.style.visibility = "hidden", e.then(() => {
				this.style.visibility = "";
			}));
		}
		disconnectedCallback() {
			this.app && Yr(this.app), this.app = null;
		}
	});
}
function ra(e, t) {
	let n = tn({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Qt(r);
		}
	}), r = /* @__PURE__ */ N(0);
	wn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Qt(r);
			}, 1e3);
		});
		return () => {
			i(), clearTimeout(n);
		};
	});
	let i = !1, a = !1, o = "", s = (e, t) => {
		i = !0, a = !1;
		let r = (r) => {
			i = !1, t === o && (Object.assign(n, r), a && s(e, t));
		};
		e.get(t).then((e) => r({
			data: e,
			error: "",
			loading: !1
		}), (e) => r({
			error: String(e?.message || e),
			loading: !1
		}));
	};
	return wn(() => {
		let n = e(), c = t();
		J(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function ia(e) {
	let t = tn({ ...e()?.params?.() ?? {} });
	return wn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
function aa(e, t) {
	let n = t.api.mount(e, t.slot, t.target, t.context);
	return {
		update(r) {
			n && JSON.stringify(r.target) === JSON.stringify(t.target) ? n.context(r.context) : (n?.destroy(), n = r.api.mount(e, r.slot, r.target, r.context)), t = r;
		},
		destroy() {
			n?.destroy();
		}
	};
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/@lucide/svelte/dist/utils/defaultAttributes.js
var oa = {
	xmlns: "http://www.w3.org/2000/svg",
	width: 24,
	height: 24,
	viewBox: "0 0 24 24",
	fill: "none",
	stroke: "currentColor",
	"stroke-width": 2,
	"stroke-linecap": "round",
	"stroke-linejoin": "round"
}, sa = (...e) => e.filter((e, t, n) => !!e && e.trim() !== "" && n.indexOf(e) === t).join(" ").trim();
// node_modules/@lucide/svelte/dist/utils/buildLucideIconNode.js
function ca(e) {
	return e != null;
}
function la(e, t = {}) {
	let n = t.attributeNames ?? {}, r = (e) => n[e] ?? e, i = e.size ?? e.width ?? oa.width, a = e.size ?? e.height ?? oa.height, o = e.aliases?.filter((e) => typeof e == "string" && e.trim() !== "").map((e) => `lucide-${e}`) ?? [], s = [...e.name ? [`lucide-${e.name}`] : [], ...o], c = t.className?.split(" ").filter(Boolean) ?? [], l = t.includeDefaultClasses === !1 ? sa(...c) : sa("lucide", ...s, ...c), u = t.absoluteStrokeWidth ? Number(t.strokeWidth ?? oa["stroke-width"]) * Number(e.size ?? e.width ?? oa.width) / Number(t.size ?? t.width ?? oa.width) : t.strokeWidth ?? oa["stroke-width"];
	return [
		"svg",
		{
			...Object.entries(oa).reduce((e, [t, n]) => (e[r(t)] = n, e), {}),
			..."color" in t && t.color && { [r("stroke")]: t.color },
			..."size" in t && ca(t.size) && {
				[r("width")]: t.size,
				[r("height")]: t.size
			},
			..."width" in t && ca(t.width) && { [r("width")]: t.width },
			..."height" in t && ca(t.height) && { [r("height")]: t.height },
			[r("stroke-width")]: u,
			...l && { [r("class")]: l },
			[r("viewBox")]: `0 0 ${i} ${a}`,
			...t.hasA11yProp === !1 ? { [r("aria-hidden")]: "true" } : {},
			..."attributes" in t && t.attributes
		},
		e.node.map((e) => {
			let [n, i, a] = e, o = t.nonScalingStroke ? {
				[r("vector-effect")]: "non-scaling-stroke",
				...i
			} : i;
			return a ? [
				n,
				o,
				a
			] : [n, o];
		})
	];
}
// node_modules/@lucide/svelte/dist/utils/hasA11yProp.js
var ua = (e) => {
	for (let t in e) if (t.startsWith("aria-") || t === "role" || t === "title") return !0;
	return !1;
}, da = Symbol("lucide-context"), fa = () => Qe(da), pa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy",
	"color",
	"size",
	"width",
	"height",
	"strokeWidth",
	"absoluteStrokeWidth",
	"nonScalingStroke",
	"iconNode",
	"icon",
	"class",
	"children"
]), ma = /* @__PURE__ */ Or("<svg><!><!></svg>");
function ha(e, t) {
	$e(t, !0);
	let n = fa() ?? {}, r = ta(t, "color", 19, () => n.color ?? "currentColor"), i = ta(t, "size", 19, () => n.size ?? 24), a = ta(t, "width", 19, i), o = ta(t, "height", 19, i), s = ta(t, "strokeWidth", 19, () => n.strokeWidth ?? 2), c = ta(t, "absoluteStrokeWidth", 19, () => n.absoluteStrokeWidth ?? !1), l = ta(t, "nonScalingStroke", 19, () => n.nonScalingStroke ?? !1), u = ta(t, "iconNode", 19, () => []), d = ta(t, "icon", 19, () => ({
		node: u(),
		aliases: [],
		size: 24
	})), f = /* @__PURE__ */ $(t, pa), p = /* @__PURE__ */ j(() => !!t.children || ua(f)), m = /* @__PURE__ */ j(() => la(d(), {
		color: r(),
		width: a(),
		height: o(),
		strokeWidth: s(),
		absoluteStrokeWidth: c(),
		nonScalingStroke: l(),
		className: sa("lucide-icon", n.class),
		hasA11yProp: J(p),
		attributes: f
	})), h = /* @__PURE__ */ j(() => b(J(m), 3)), _ = /* @__PURE__ */ j(() => J(h)[1]), v = /* @__PURE__ */ j(() => y(J(h)[2], () => [], !0)), x = /* @__PURE__ */ j(() => ({
		...J(_),
		class: [...J(_).class.split(" "), t.class]
	}));
	var S = ma();
	Wi(S, () => ({ ...J(x) }));
	var C = L(S);
	ai(C, 17, () => J(v), ti, (e, t) => {
		var n = /* @__PURE__ */ j(() => b(J(t), 2));
		let r = () => J(n)[0], i = () => J(n)[1];
		var a = kr();
		fi(dn(a), r, !0, (e, t) => {
			Wi(e, () => ({ ...i() }));
		}), X(e, a);
	}), Zr(R(C), () => t.children ?? g), O(S), X(e, S), et();
}
// node_modules/@lucide/svelte/dist/icons/inbox.svelte
var ga = {
	name: "inbox",
	size: 24,
	node: [["polyline", { points: "22 12 16 12 14 15 10 15 8 12 2 12" }], ["path", { d: "M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" }]]
};
ga.node;
var _a = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function va(e, t) {
	let n = /* @__PURE__ */ $(t, _a);
	ha(e, ea(() => n, { get icon() {
		return ga;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square.svelte
var ya = {
	name: "message-square",
	size: 24,
	node: [["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }]]
};
ya.node;
var ba = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function xa(e, t) {
	let n = /* @__PURE__ */ $(t, ba);
	ha(e, ea(() => n, { get icon() {
		return ya;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/list-checks.svelte
var Sa = {
	name: "list-checks",
	size: 24,
	node: [
		["path", { d: "M13 5h8" }],
		["path", { d: "M13 12h8" }],
		["path", { d: "M13 19h8" }],
		["path", { d: "m3 17 2 2 4-4" }],
		["path", { d: "m3 7 2 2 4-4" }]
	]
};
Sa.node;
var Ca = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function wa(e, t) {
	let n = /* @__PURE__ */ $(t, Ca);
	ha(e, ea(() => n, { get icon() {
		return Sa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/activity.svelte
var Ta = {
	name: "activity",
	size: 24,
	node: [["path", { d: "M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" }]]
};
Ta.node;
var Ea = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Da(e, t) {
	let n = /* @__PURE__ */ $(t, Ea);
	ha(e, ea(() => n, { get icon() {
		return Ta;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/terminal.svelte
var Oa = {
	name: "terminal",
	size: 24,
	node: [["path", { d: "M12 19h8" }], ["path", { d: "m4 17 6-6-6-6" }]]
};
Oa.node;
var ka = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Aa(e, t) {
	let n = /* @__PURE__ */ $(t, ka);
	ha(e, ea(() => n, { get icon() {
		return Oa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/boxes.svelte
var ja = {
	name: "boxes",
	size: 24,
	node: [
		["path", { d: "M2.97 12.92A2 2 0 0 0 2 14.63v3.24a2 2 0 0 0 .97 1.71l3 1.8a2 2 0 0 0 2.06 0L12 19v-5.5l-5-3-4.03 2.42Z" }],
		["path", { d: "m7 16.5-4.74-2.85" }],
		["path", { d: "m7 16.5 5-3" }],
		["path", { d: "M7 16.5v5.17" }],
		["path", { d: "M12 13.5V19l3.97 2.38a2 2 0 0 0 2.06 0l3-1.8a2 2 0 0 0 .97-1.71v-3.24a2 2 0 0 0-.97-1.71L17 10.5l-5 3Z" }],
		["path", { d: "m17 16.5-5-3" }],
		["path", { d: "m17 16.5 4.74-2.85" }],
		["path", { d: "M17 16.5v5.17" }],
		["path", { d: "M7.97 4.42A2 2 0 0 0 7 6.13v4.37l5 3 5-3V6.13a2 2 0 0 0-.97-1.71l-3-1.8a2 2 0 0 0-2.06 0l-3 1.8Z" }],
		["path", { d: "M12 8 7.26 5.15" }],
		["path", { d: "m12 8 4.74-2.85" }],
		["path", { d: "M12 13.5V8" }]
	]
};
ja.node;
var Ma = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Na(e, t) {
	let n = /* @__PURE__ */ $(t, Ma);
	ha(e, ea(() => n, { get icon() {
		return ja;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/rocket.svelte
var Pa = {
	name: "rocket",
	size: 24,
	node: [
		["path", { d: "M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" }],
		["path", { d: "M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09" }],
		["path", { d: "M9 12a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.4 22.4 0 0 1-4 2z" }],
		["path", { d: "M9 12H4s.55-3.03 2-4c1.62-1.08 5 .05 5 .05" }]
	]
};
Pa.node;
var Fa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ia(e, t) {
	let n = /* @__PURE__ */ $(t, Fa);
	ha(e, ea(() => n, { get icon() {
		return Pa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/sun.svelte
var La = {
	name: "sun",
	size: 24,
	node: [
		["circle", {
			cx: "12",
			cy: "12",
			r: "4"
		}],
		["path", { d: "M12 2v2" }],
		["path", { d: "M12 20v2" }],
		["path", { d: "m4.93 4.93 1.41 1.41" }],
		["path", { d: "m17.66 17.66 1.41 1.41" }],
		["path", { d: "M2 12h2" }],
		["path", { d: "M20 12h2" }],
		["path", { d: "m6.34 17.66-1.41 1.41" }],
		["path", { d: "m19.07 4.93-1.41 1.41" }]
	]
};
La.node;
var Ra = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function za(e, t) {
	let n = /* @__PURE__ */ $(t, Ra);
	ha(e, ea(() => n, { get icon() {
		return La;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/moon.svelte
var Ba = {
	name: "moon",
	size: 24,
	node: [["path", { d: "M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401" }]]
};
Ba.node;
var Va = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ha(e, t) {
	let n = /* @__PURE__ */ $(t, Va);
	ha(e, ea(() => n, { get icon() {
		return Ba;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/puzzle.svelte
var Ua = {
	name: "puzzle",
	size: 24,
	node: [["path", { d: "M15.39 4.39a1 1 0 0 0 1.68-.474 2.5 2.5 0 1 1 3.014 3.015 1 1 0 0 0-.474 1.68l1.683 1.682a2.414 2.414 0 0 1 0 3.414L19.61 15.39a1 1 0 0 1-1.68-.474 2.5 2.5 0 1 0-3.014 3.015 1 1 0 0 1 .474 1.68l-1.683 1.682a2.414 2.414 0 0 1-3.414 0L8.61 19.61a1 1 0 0 0-1.68.474 2.5 2.5 0 1 1-3.014-3.015 1 1 0 0 0 .474-1.68l-1.683-1.682a2.414 2.414 0 0 1 0-3.414L4.39 8.61a1 1 0 0 1 1.68.474 2.5 2.5 0 1 0 3.014-3.015 1 1 0 0 1-.474-1.68l1.683-1.682a2.414 2.414 0 0 1 3.414 0z" }]]
};
Ua.node;
var Wa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ga(e, t) {
	let n = /* @__PURE__ */ $(t, Wa);
	ha(e, ea(() => n, { get icon() {
		return Ua;
	} }));
}
// plugins/desk/panel.svelte.js
var Ka = 300, qa = 760, Ja = 212, Ya = 360, Xa = 440;
function Za(e) {
	let t = window.innerWidth - Ja - Ya;
	return Math.round(Math.max(Ka, Math.min(qa, t, e)));
}
var Qa = Number(localStorage.getItem("desk-panel-w")), $a = window.matchMedia("(max-width: 768px)"), eo = tn({
	w: Za(Qa > 0 ? Qa : Xa),
	mobile: $a.matches
});
$a.addEventListener("change", (e) => {
	eo.mobile = e.matches;
}), window.addEventListener("resize", () => {
	eo.w = Za(eo.w);
});
function to(e) {
	eo.w = Za(e), localStorage.setItem("desk-panel-w", String(eo.w));
}
function no() {
	to(Xa);
}
// kit/format.js
function ro(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
// node_modules/@lucide/svelte/dist/icons/x.svelte
var io = {
	name: "x",
	size: 24,
	node: [["path", { d: "M18 6 6 18" }], ["path", { d: "m6 6 12 12" }]]
};
io.node;
var ao = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function oo(e, t) {
	let n = /* @__PURE__ */ $(t, ao);
	ha(e, ea(() => n, { get icon() {
		return io;
	} }));
}
// plugins/desk/Drawer.svelte
var so = /* @__PURE__ */ Y("<div class=\"err svelte-gsmuzk\"> </div>"), co = /* @__PURE__ */ Y("<div class=\"plugin-host\"></div>"), lo = /* @__PURE__ */ Y("<div class=\"row actions svelte-gsmuzk\" data-testid=\"drawer-actions\"></div>"), uo = /* @__PURE__ */ Y("<h3 class=\"svelte-gsmuzk\"> </h3>"), fo = /* @__PURE__ */ Y("<!> <div data-testid=\"drawer-plugin\"><div class=\"plugin-host\"></div></div>", 1), po = /* @__PURE__ */ Y("<p class=\"faint\">No drawer plugins are enabled.</p>"), mo = /* @__PURE__ */ Y("<!> <!>", 1), ho = /* @__PURE__ */ Y("<p class=\"faint\">Loading…</p>"), go = /* @__PURE__ */ Y("<div class=\"backdrop svelte-gsmuzk\" role=\"presentation\"></div> <aside class=\"drawer svelte-gsmuzk\" data-testid=\"drawer\" aria-label=\"ticket\"><header class=\"svelte-gsmuzk\"><span class=\"mono muted id svelte-gsmuzk\"> </span> <button class=\"ghost icon\" aria-label=\"close\"><!></button></header> <div class=\"scrollbody svelte-gsmuzk\"><!></div></aside>", 1);
function _o(e, t) {
	$e(t, !0);
	let n = ra(() => t.api, () => "api/ticket/" + encodeURIComponent(t.id)), r = () => t.api.go({ ticket: "" }), i = (e, t) => ({
		kind: e.kind,
		label: e.labels ?? [],
		pool: e.pool || "",
		workflow: ro(t?.ticket ?? e).workflow || ""
	}), a = (e, n, r) => t.api.contributions(e, i(n, r)).sort((e, t) => Number(t.source === "builtin") - Number(e.source === "builtin")), o = (e) => qe({
		ticket: e.ticket,
		run: e.run,
		trace: e.trace,
		comments: e.comments,
		sessions: e.sessions
	});
	var s = go();
	_r("keydown", an, (e) => {
		e.key === "Escape" && (e.preventDefault(), r());
	});
	var c = dn(s), l = R(c, 2), u = L(l), d = L(u), f = fn(d, !0), p = R(d, 2);
	oo(L(p), { size: 14 }), O(p), O(u);
	var m = R(u, 2), h = L(m), g = (e) => {
		var t = so(), r = fn(t, !0);
		z(() => Wr(r, n.error)), X(e, t);
	}, _ = (e) => {
		let r = /* @__PURE__ */ j(() => n.data.ticket), i = /* @__PURE__ */ j(() => o(n.data)), s = /* @__PURE__ */ j(() => a("action", J(r), n.data.run));
		var c = mo(), l = dn(c), u = (e) => {
			var n = lo();
			ai(n, 21, () => J(s), (e) => e.plugin + e.element, (e, n) => {
				var r = co();
				pi(r, (e, t) => aa?.(e, t), () => ({
					api: t.api,
					slot: "action",
					target: {
						plugin: J(n).plugin,
						element: J(n).element
					},
					context: J(i)
				})), z(() => Q(r, "data-plugin", J(n).plugin)), X(e, r);
			}), O(n), X(e, n);
		};
		Z(l, (e) => {
			J(s).length && e(u);
		}), ai(R(l, 2), 17, () => a("drawer", J(r), n.data.run), (e) => e.plugin + e.element, (e, n) => {
			var r = fo(), a = dn(r), o = (e) => {
				var t = uo(), r = fn(t, !0);
				z(() => Wr(r, J(n).title)), X(e, t);
			};
			Z(a, (e) => {
				J(n).title && e(o);
			});
			var s = R(a, 2), c = L(s);
			pi(c, (e, t) => aa?.(e, t), () => ({
				api: t.api,
				slot: "drawer",
				target: {
					plugin: J(n).plugin,
					element: J(n).element
				},
				context: J(i)
			})), O(s), z(() => {
				Q(s, "data-plugin", J(n).plugin), Q(c, "data-plugin", J(n).plugin);
			}), X(e, r);
		}, (e) => {
			X(e, po());
		}), X(e, c);
	}, v = (e) => {
		X(e, ho());
	};
	Z(h, (e) => {
		n.error ? e(g) : n.data ? e(_, 1) : e(v, -1);
	}), O(m), O(l), z(() => Wr(f, t.id)), vr("click", c, r), vr("click", p, r), X(e, s), et();
}
yr(["click"]);
// plugins/desk/SidePanel.svelte
var vo = /* @__PURE__ */ Y("<div class=\"grip svelte-4g1rs0\" role=\"slider\" aria-orientation=\"horizontal\" aria-label=\"panel width\" aria-valuemin=\"300\" aria-valuemax=\"760\" tabindex=\"0\"></div>"), yo = /* @__PURE__ */ Y("<div class=\"plugin-host\"></div>"), bo = /* @__PURE__ */ Y("<p class=\"faint none svelte-4g1rs0\">No plugin shows sessions (slot <span class=\"mono\">session</span>).</p>"), xo = /* @__PURE__ */ Y("<aside class=\"panel svelte-4g1rs0\" tabindex=\"-1\" data-testid=\"session-panel\" aria-label=\"session\"><!> <!></aside>");
function So(e, t) {
	$e(t, !0);
	let n = /* @__PURE__ */ j(() => t.api.contributions("session").at(-1));
	wn(() => {
		if (!eo.mobile) return;
		let e = document.body.style.overflow;
		return document.body.style.overflow = "hidden", () => {
			document.body.style.overflow = e;
		};
	});
	let r = /* @__PURE__ */ N(void 0);
	wn(() => {
		J(r)?.focus({ preventScroll: !0 });
	});
	function i(e) {
		let t = e.currentTarget;
		t.setPointerCapture(e.pointerId);
		let n = (e) => to(window.innerWidth - e.clientX), r = () => {
			t.removeEventListener("pointermove", n), t.removeEventListener("pointerup", r), t.removeEventListener("pointercancel", r);
		};
		t.addEventListener("pointermove", n), t.addEventListener("pointerup", r), t.addEventListener("pointercancel", r);
	}
	function a(e) {
		e.key === "ArrowLeft" ? (e.preventDefault(), to(eo.w + 24)) : e.key === "ArrowRight" && (e.preventDefault(), to(eo.w - 24));
	}
	var o = xo(), s = L(o), c = (e) => {
		var t = vo();
		z(() => Q(t, "aria-valuenow", eo.w)), vr("pointerdown", t, i), vr("dblclick", t, function(...e) {
			no?.apply(this, e);
		}), vr("keydown", t, a), X(e, t);
	};
	Z(s, (e) => {
		eo.mobile || e(c);
	});
	var l = R(s, 2), u = (e) => {
		var r = kr();
		ei(dn(r), () => t.id + t.item + t.readonly, (e) => {
			var r = yo();
			pi(r, (e, t) => aa?.(e, t), () => ({
				api: t.api,
				slot: "session",
				target: {
					plugin: J(n).plugin,
					element: J(n).element
				},
				context: {
					session: t.id,
					item: t.item,
					readonly: t.readonly
				}
			})), z(() => Q(r, "data-plugin", J(n).plugin)), X(e, r);
		}), X(e, r);
	}, d = (e) => {
		X(e, bo());
	};
	Z(l, (e) => {
		J(n) ? e(u) : e(d, -1);
	}), O(o), Yi(o, (e) => P(r, e), () => J(r)), X(e, o), et();
}
yr([
	"pointerdown",
	"dblclick",
	"keydown"
]);
// plugins/desk/Frame.svelte
var Co = /* @__PURE__ */ Y("<div class=\"user faint svelte-e6uu72\" title=\"signed in through the proxy\"> </div>"), wo = /* @__PURE__ */ Y("<span class=\"tabdot svelte-e6uu72\"></span>"), To = /* @__PURE__ */ Y("<span class=\"badge svelte-e6uu72\"> </span>"), Eo = /* @__PURE__ */ Y("<a><span class=\"ico svelte-e6uu72\"><!><!></span> <span class=\"lbl svelte-e6uu72\"> </span> <!></a>"), Do = /* @__PURE__ */ Y("<div class=\"hdr plugin-host svelte-e6uu72\"></div>"), Oo = /* @__PURE__ */ Y("<div class=\"plugin-host\"></div>"), ko = /* @__PURE__ */ Y("<p class=\"faint pad\" data-testid=\"view-missing\">No plugin shows <span class=\"mono\"> </span>.</p>"), Ao = /* @__PURE__ */ Y("<div data-testid=\"desk-app\"><div class=\"brand svelte-e6uu72\"><span class=\"mark svelte-e6uu72\"></span>Jobshop <span class=\"faint\">Desk</span></div> <nav aria-label=\"views\" class=\"svelte-e6uu72\"><!> <!> <!></nav> <div class=\"foot svelte-e6uu72\"><span class=\"live svelte-e6uu72\"><span></span> <span class=\"faint\"> </span></span> <span class=\"tools svelte-e6uu72\"><button class=\"ghost icon svelte-e6uu72\" aria-label=\"toggle theme\" title=\"toggle theme\"><!></button></span></div> <main class=\"svelte-e6uu72\"><!></main> <!> <!></div>");
function jo(e, t) {
	$e(t, !0);
	let n = ta(t, "api", 7), r = ia(() => n()), i = /* @__PURE__ */ N(""), a = /* @__PURE__ */ N(!1), o = /* @__PURE__ */ N(tn({})), s = /* @__PURE__ */ N(!1);
	Qr(() => {
		P(s, n().theme === "light"), P(i, n().user, !0), J(i) || n().get("api/whoami").then((e) => P(i, e.user || "", !0)).catch(() => {});
		let e = n().live((e) => P(a, e, !0)), t = n().badges((e) => P(o, e, !0));
		return () => {
			e(), t();
		};
	});
	function c() {
		P(s, !J(s)), n().theme = J(s) ? "light" : "dark";
	}
	let l = {
		attention: {
			icon: va,
			key: "a"
		},
		chat: {
			icon: xa,
			key: "c"
		},
		work: {
			icon: wa,
			key: "w"
		},
		activity: {
			icon: Da,
			key: "y"
		},
		sessions: {
			icon: Aa,
			key: "s"
		},
		pools: {
			icon: Na,
			key: "p"
		},
		deploys: {
			icon: Ia,
			key: "d"
		}
	}, u = Object.keys(l), d = (e) => {
		let t = u.indexOf(e);
		return t < 0 ? u.length : t;
	}, f = /* @__PURE__ */ j(() => n().contributions("view")), p = /* @__PURE__ */ j(() => J(f).filter((e, t) => e.id && J(f).findIndex((t) => t.id === e.id) === t).map((e, t) => ({
		id: e.id ?? "",
		title: e.title || e.id || "",
		i: t
	})).sort((e, t) => d(e.id) - d(t.id) || e.i - t.i).map((e) => ({
		...e,
		icon: l[e.id]?.icon || Ga,
		key: l[e.id]?.key || ""
	}))), m = /* @__PURE__ */ j(() => n().contributions("header")), h = /* @__PURE__ */ j(() => r.view || "attention"), g = /* @__PURE__ */ j(() => J(f).find((e) => e.id === J(h))), _ = (e) => n().go(e), v = 0;
	function y(e) {
		if (e.key === "Escape" && r.session && !r.ticket && !e.defaultPrevented) {
			n().closeSession();
			return;
		}
		let t = e.target;
		if (!(e.metaKey || e.ctrlKey || e.altKey || t.closest("input, textarea, select, [contenteditable]"))) {
			if (v && Date.now() - v < 1200) {
				v = 0;
				let t = J(p).find((t) => t.key && t.key === e.key);
				t && (e.preventDefault(), _({
					view: t.id,
					ticket: ""
				}));
			} else e.key === "g" && (v = Date.now());
		}
	}
	var b = Ao();
	_r("keydown", an, y);
	let x, S;
	var C = R(L(b), 2), w = L(C), ee = (e) => {
		var t = Co(), n = fn(t, !0);
		z(() => Wr(n, J(i))), X(e, t);
	};
	Z(w, (e) => {
		J(i) && e(ee);
	});
	var te = R(w, 2);
	ai(te, 17, () => J(p), (e) => e.id, (e, t) => {
		let n = /* @__PURE__ */ j(() => J(t).icon);
		var r = Eo();
		let i;
		var a = L(r), s = L(a);
		di(s, () => J(n), (e, t) => {
			t(e, { size: 14 });
		});
		var c = R(s), l = (e) => {
			X(e, wo());
		};
		Z(c, (e) => {
			(J(o)[J(t).id] ?? 0) > 0 && e(l);
		}), O(a);
		var u = R(a, 2), d = fn(u, !0), f = R(u, 2), p = (e) => {
			var n = To(), r = fn(n, !0);
			z(() => Wr(r, J(o)[J(t).id])), X(e, n);
		};
		Z(f, (e) => {
			(J(o)[J(t).id] ?? 0) > 0 && e(p);
		}), O(r), z(() => {
			Q(r, "href", "?view=" + J(t).id), Q(r, "data-view", J(t).id), Q(r, "title", "Go to " + J(t).title + (J(t).key ? " (g " + J(t).key + ")" : "")), i = Ci(r, 1, "svelte-e6uu72", null, i, { active: J(h) === J(t).id }), Wr(d, J(t).title);
		}), vr("click", r, (e) => {
			e.preventDefault(), _({
				view: J(t).id,
				ticket: ""
			});
		}), X(e, r);
	}), ai(R(te, 2), 17, () => J(m), (e) => e.plugin + e.element, (e, t) => {
		var r = Do();
		pi(r, (e, t) => aa?.(e, t), () => ({
			api: n(),
			slot: "header",
			target: {
				plugin: J(t).plugin,
				element: J(t).element
			}
		})), z(() => Q(r, "data-plugin", J(t).plugin)), X(e, r);
	}), O(C);
	var ne = R(C, 2), re = L(ne), ie = L(re);
	let ae;
	var oe = fn(R(ie, 2), !0);
	O(re);
	var se = R(re, 2), ce = L(se), le = L(ce), ue = (e) => {
		Ha(e, { size: 14 });
	}, de = (e) => {
		za(e, { size: 14 });
	};
	Z(le, (e) => {
		J(s) ? e(ue) : e(de, -1);
	}), O(ce), O(se), O(ne);
	var fe = R(ne, 2), pe = L(fe), me = (e) => {
		var t = kr();
		ei(dn(t), () => J(g).plugin + J(g).element, (e) => {
			var t = Oo();
			pi(t, (e, t) => aa?.(e, t), () => ({
				api: n(),
				slot: "view",
				target: {
					plugin: J(g).plugin,
					element: J(g).element
				}
			})), z(() => Q(t, "data-plugin", J(g).plugin)), X(e, t);
		}), X(e, t);
	}, he = (e) => {
		var t = ko(), n = fn(R(L(t)), !0);
		Oe(), O(t), z(() => Wr(n, J(h))), X(e, t);
	};
	Z(pe, (e) => {
		J(g) ? e(me) : e(he, -1);
	}), O(fe);
	var ge = R(fe, 2), _e = (e) => {
		var t = kr();
		ei(dn(t), () => r.session, (e) => {
			{
				let t = /* @__PURE__ */ j(() => r.item || ""), i = /* @__PURE__ */ j(() => r.ro === "1");
				So(e, {
					get api() {
						return n();
					},
					get id() {
						return r.session;
					},
					get item() {
						return J(t);
					},
					get readonly() {
						return J(i);
					}
				});
			}
		}), X(e, t);
	};
	Z(ge, (e) => {
		r.session && e(_e);
	});
	var ve = R(ge, 2), ye = (e) => {
		_o(e, {
			get api() {
				return n();
			},
			get id() {
				return r.ticket;
			}
		});
	};
	Z(ve, (e) => {
		r.ticket && e(ye);
	}), O(b), z(() => {
		x = Ci(b, 1, "shell svelte-e6uu72", null, x, { withpanel: !!r.session }), S = Ti(b, "", S, { "--pw": eo.w + "px" }), Q(re, "title", J(a) ? "live" : "disconnected"), ae = Ci(ie, 1, "dot", null, ae, {
			ok: J(a),
			bad: !J(a)
		}), Wr(oe, J(a) ? "live" : "offline");
	}), vr("click", ce, c), X(e, b), et();
}
// plugins/desk/index.js
yr(["click"]), na("desk-app", jo);
