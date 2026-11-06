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
	return D(/* @__PURE__ */ dn(E));
}
function O(t) {
	if (T) {
		if (/* @__PURE__ */ dn(E) !== null) throw Ce(), e;
		E = t;
	}
}
function Oe(e = 1) {
	if (T) {
		for (var t = e, n = E; t--;) n = /* @__PURE__ */ dn(n);
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
		var i = /* @__PURE__ */ dn(n);
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
		r: U,
		l: null
	};
}
function et(e) {
	var t = k, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) En(r);
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
	if (rt.length === 0 && !kt) {
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
	var t = V, n = U;
	H(null), qn(null);
	try {
		return e();
	} finally {
		H(t), qn(n);
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
	var s = U, c = mt(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				yn(e, s);
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
		Promise.all(n.map((e) => /* @__PURE__ */ yt(e))).then(u).catch((e) => yn(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), ht());
	}) : f();
}
function mt() {
	var e = U, t = V, n = k, r = M;
	return function(i = !0) {
		qn(e), H(t), Ze(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function ht(e = !0) {
	qn(null), H(null), Ze(null), e && M?.deactivate();
}
function gt() {
	var e = U, t = e.b, n = M, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function _t(e) {
	var n = 2 | C;
	return U !== null && (U.f |= ae), {
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
		parent: U,
		ac: null
	};
}
var vt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function yt(e, n, r) {
	let i = U;
	i === null && Fe();
	var a = void 0, o = Gt(t), s = !V, c = /* @__PURE__ */ new Set();
	return kn(() => {
		var t = U, n = v();
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
			l?.(), c.delete(n), t !== vt && (r.activate(), t ? (o.f |= ue, Zt(o, t)) : (o.f & 8388608 && (o.f ^= ue), Zt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), wn(() => {
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
	return Yn(t), t;
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
		for (var n = 0; n < t.length; n += 1) B(t[n]);
	}
}
function St(e) {
	var n, r = U, i = e.parent;
	if (!Wn && i !== null && e.v !== t && i.f & 24576) return Se(), e.v;
	qn(i);
	try {
		xt(e), n = ir(e);
	} finally {
		qn(r);
	}
	return n;
}
function Ct(e) {
	var t = St(e);
	!e.equals(t) && (e.wv = tr(), (!M?.is_fork || e.deps === null) && (M === null ? e.v = t : (M.capture(e, t, !0), Dt?.capture(e, t, !0)), e.deps === null)) ? A(e, S) : Wn || (N === null ? st(e) : (Cn() || M?.is_fork) && N.set(e, t));
}
function wt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && ft(() => {
		t.ac.abort(be), t.ac = null;
	}), t.fn !== null && (t.teardown = g), sr(t, 0), Pn(t));
}
function Tt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && cr(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var Et = null, M = null, Dt = null, N = null, Ot = null, kt = !1, At = !1, jt = null, Mt = null, Nt = 0, Pt = 1, Ft = class e {
	id = Pt++;
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
		for (var t = jt = [], n = [], r = Mt = []; this.#a.length > 0;) {
			Nt++ > 1e3 && (this.#b(), It());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Vt(e), this.#d() || this.discard(), t;
			}
		}
		if (M = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (jt = null, Mt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Bt(e, t);
			r.length > 0 && M.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), Dt = this, Rt(n), Rt(t), Dt = null, this.#i?.resolve();
			var o = M;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ut.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= S;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= S : i & 4 ? t.push(r) : nr(r) && (i & 16 && this.add_dirty_reaction(r, w), cr(r));
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
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), N?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		M = this;
	}
	deactivate() {
		M = null, N = null;
	}
	flush() {
		try {
			At = !0, M = this, this.#m();
		} finally {
			Nt = 0, Ot = null, jt = null, Mt = null, At = !1, M = null, N = null, Ut.clear();
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
			!At && at(() => {
				t.started || t.flush();
			});
		}
		return M;
	}
	apply() {
		N = null;
	}
	schedule(e) {
		Ot = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? Et = e : t.prev = e, this.linked = !1;
		}
	}
};
function It() {
	try {
		Be();
	} catch (e) {
		yn(e, Ot);
	}
}
var Lt = null;
function Rt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && nr(r) && (Lt = /* @__PURE__ */ new Set(), cr(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && Ln(r), Lt?.size > 0)) {
				Ut.clear();
				for (let e of Lt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Lt.has(n) && (Lt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || cr(n);
					}
				}
				Lt.clear();
			}
		}
		Lt = null;
	}
}
function zt(e) {
	M.schedule(e);
}
function Bt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), A(e, S);
		for (var n = e.first; n !== null;) Bt(n, t), n = n.next;
	}
}
function Vt(e) {
	A(e, S);
	for (var t = e.first; t !== null;) Vt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Ht = /* @__PURE__ */ new Set(), Ut = /* @__PURE__ */ new Map(), Wt = !1;
function Gt(e, t) {
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
function Kt(e, t) {
	let n = Gt(e, t);
	return Yn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function qt(e, t = !1, n = !0) {
	let r = Gt(e);
	return t || (r.equals = Ne), r;
}
function Jt(e, t, n = !1) {
	return V !== null && (!Kn || V.f & 131072) && nt() && V.f & 4325394 && (Jn === null || !Jn.has(e)) && We(), Zt(e, n ? nn(t) : t, Mt);
}
var Yt = null, Xt = 0;
function Zt(e, t, n = null) {
	if (!e.equals(t)) {
		Wn ? Ut.set(e, t) : Ut.has(e) || Ut.set(e, e.v);
		var r = Ft.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && St(t), N === null && st(t);
		}
		e.wv = tr(), Yt = null, Xt = 0, tn(e, C, n), Yt = null, nt() && U !== null && U.f & 1024 && !(U.f & 96) && (K === null ? Xn([e]) : K.push(e)), !r.is_fork && Ht.size > 0 && !Wt && Qt();
	}
	return t;
}
function Qt() {
	Wt = !1;
	for (let e of Ht) {
		e.f & 1024 && A(e, w);
		let t;
		try {
			t = nr(e);
		} catch {
			t = !0;
		}
		t && cr(e);
	}
	Ht.clear();
}
function $t(e, t = 1) {
	var n = q(e), r = t === 1 ? n++ : n--;
	return Jt(e, n), r;
}
function en(e) {
	Jt(e, e.v + 1);
}
function tn(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = nt(), a = r.length;
		if (Xt += a, Xt > 1e5 && Yt === null && (Yt = /* @__PURE__ */ new Set()), Yt !== null) {
			if (Yt.has(e)) return;
			Yt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & C) === 0;
				if (l && A(s, t), c & 131072) Ht.add(s);
				else if (c & 2) {
					var u = s;
					N?.delete(u), tn(u, w, n);
				} else if (l) {
					var d = s;
					c & 16 && Lt !== null && Lt.add(d), n === null ? zt(d) : n.push(d);
				}
			}
		}
	}
}
function nn(e) {
	if (typeof e != "object" || !e || de in e || fe in e) return e;
	let n = p(e);
	if (n !== d && n !== f) return e;
	var r = /* @__PURE__ */ new Map(), a = i(e), o = /* @__PURE__ */ Kt(0), s = null, c = $n, u = (e) => {
		if ($n === c) return e();
		var t = V, n = $n;
		H(null), er(c);
		var r = e();
		return H(t), er(n), r;
	};
	return a && r.set("length", /* @__PURE__ */ Kt(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || He();
			var i = r.get(t);
			return i === void 0 ? u(() => {
				var e = /* @__PURE__ */ Kt(n.value, s);
				return r.set(t, e), e;
			}) : Jt(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = u(() => /* @__PURE__ */ Kt(t, s));
					r.set(n, e), en(o);
				}
			} else Jt(i, t), en(o);
			return !0;
		},
		get(n, i, a) {
			if (i === de) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || l(n, i)?.writable) && (o = u(() => /* @__PURE__ */ Kt(nn(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = q(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = q(a);
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
			return (i !== void 0 || (U !== null || V !== null) && (!a || l(e, n)?.writable)) && (i === void 0 && (i = u(() => /* @__PURE__ */ Kt(a ? nn(e[n]) : t, s)), r.set(n, i)), q(i) === t) ? !1 : a;
		},
		set(e, n, i, c) {
			var d = r.get(n), f = n in e;
			if (a && n === "length") for (var p = i; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = u(() => /* @__PURE__ */ Kt(t, s)), r.set(p + "", m)) : Jt(m, t);
			}
			if (d === void 0) (!f || l(e, n)?.writable) && (d = u(() => /* @__PURE__ */ Kt(void 0, s)), Jt(d, nn(i)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = u(() => nn(i));
				Jt(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, i), !f) {
				if (a && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && Jt(_, v + 1);
				}
				en(o);
			}
			return !0;
		},
		ownKeys(e) {
			q(o);
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
function rn(e) {
	try {
		if (typeof e == "object" && e && de in e) return e[de];
	} catch {}
	return e;
}
function an(e, t) {
	return Object.is(rn(e), rn(t));
}
var on, sn, cn, ln;
function un() {
	if (on === void 0) {
		on = window, sn = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		cn = l(t, "firstChild").get, ln = l(t, "nextSibling").get, m(e) && (e[ge] = void 0, e[he] = null, e[_e] = void 0, e.__e = void 0), m(n) && (n[ve] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function F(e) {
	return cn.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function dn(e) {
	return ln.call(e);
}
function I(e, t) {
	if (!T) return /* @__PURE__ */ F(e);
	var n = /* @__PURE__ */ F(E);
	if (n === null) n = E.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), D(r), r;
	}
	return t && _n(n), D(n), n;
}
function fn(e, t = !1) {
	if (!T) {
		var n = /* @__PURE__ */ F(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ dn(n) : n;
	}
	if (t) {
		if (E?.nodeType !== 3) {
			var r = P();
			return E?.before(r), D(r), r;
		}
		_n(E);
	}
	return E;
}
function pn(e, t = !1) {
	if (!T) return /* @__PURE__ */ F(e);
	var n = I(e, t);
	return O(e), n;
}
function L(e, t = 1, n = !1) {
	let r = T ? E : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ dn(r);
	if (!T) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), D(a), a;
		}
		_n(r);
	}
	return D(r), r;
}
function mn(e) {
	e.textContent = "";
}
function hn() {
	return !1;
}
function gn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function _n(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function vn(e) {
	var t = U;
	if (t === null) return V.f |= ue, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	yn(e, t);
}
function yn(t, n) {
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
function bn(e) {
	U === null && (V === null && ze(e), Re()), Wn && Le(e);
}
function xn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Sn(e, t) {
	var n = U;
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
	if (e & 4) jt === null ? Ft.ensure().schedule(r) : jt.push(r);
	else if (t !== null) {
		try {
			cr(r);
		} catch (e) {
			throw B(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ie));
	}
	if (i !== null && (i.parent = n, n !== null && xn(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Cn() {
	return V !== null && !Kn;
}
function wn(e) {
	let t = Sn(8, null);
	return A(t, S), t.teardown = e, t;
}
function Tn(e) {
	bn("$effect");
	var t = U.f;
	if (!V && t & 32 && k !== null && !k.i) {
		var n = k;
		(n.e ??= []).push(e);
	} else return En(e);
}
function En(e) {
	return Sn(4 | oe, e);
}
function Dn(e) {
	Ft.ensure();
	let t = Sn(64 | ae, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? Rn(t, () => {
			B(t), n(void 0);
		}) : (B(t), n(void 0));
	});
}
function On(e) {
	return Sn(4, e);
}
function kn(e) {
	return Sn(le | ae, e);
}
function An(e, t = 0) {
	return Sn(8 | t, e);
}
function R(e, t = [], n = [], r = []) {
	pt(r, t, n, (t) => {
		Sn(8, () => {
			e(...t.map(q));
		});
	});
}
function jn(e, t = 0) {
	return Sn(16 | t, e);
}
function Mn(e, t = 0) {
	return Sn(x | t, e);
}
function z(e) {
	return Sn(32 | ae, e);
}
function Nn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Wn, r = V;
		Gn(!0), H(null);
		try {
			t.call(null);
		} catch (t) {
			yn(t, e.parent);
		} finally {
			Gn(n), H(r);
		}
	}
}
function Pn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && ft(() => {
			e.abort(be);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : B(n, t), n = r;
	}
}
function Fn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || B(t), t = n;
	}
}
function B(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (In(e.nodes.start, e.nodes.end), n = !0), e.f |= re, Pn(e, t && !n), sr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	Nn(e), e.f ^= re, e.f |= te;
	var i = e.parent;
	i !== null && i.first !== null && Ln(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function In(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ dn(e);
		e.remove(), e = n;
	}
}
function Ln(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function Rn(e, t, n = !0) {
	var r = [];
	e.f |= 256, zn(e, r, !0);
	var i = () => {
		n && B(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function zn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ee;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				zn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Bn(e) {
	e.f &= -257, Vn(e, !0);
}
function Vn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ee, e.f & 1024 || (A(e, C), Ft.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Vn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Hn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ dn(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Un = null, Wn = !1;
function Gn(e) {
	Wn = e;
}
var V = null, Kn = !1;
function H(e) {
	V = e;
}
var U = null;
function qn(e) {
	U = e;
}
var Jn = null;
function Yn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (Jn ??= /* @__PURE__ */ new Set()).add(e);
}
var W = null, G = 0, K = null;
function Xn(e) {
	K = e;
}
var Zn = 1, Qn = 0, $n = Qn;
function er(e) {
	$n = e;
}
function tr() {
	return ++Zn;
}
function nr(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (nr(a) && Ct(a), a.wv > e.wv) return !0;
		}
		t & 512 && N === null && A(e, S);
	}
	return !1;
}
function rr(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Jn !== null && Jn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? rr(a, t, !1) : t === a && (n ? A(a, C) : a.f & 1024 && A(a, w), zt(a));
	}
}
function ir(e) {
	var t = W, n = G, r = K, i = V, a = Jn, o = k, s = Kn, c = $n, l = e.f;
	W = null, G = 0, K = null, V = l & 96 ? null : e, Jn = null, Ze(e.ctx), Kn = !1, $n = ++Qn, e.ac !== null && (ft(() => {
		e.ac.abort(be);
	}), e.ac = null);
	try {
		e.f |= ce;
		var u = e.fn, d = u();
		e.f |= ne;
		var f = ar(e);
		if (nt() && K !== null && !Kn && f !== null && !(e.f & 6146)) for (var p = 0; p < K.length; p++) rr(K[p], e);
		if (i !== null && i !== e) {
			if (Qn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Qn;
			if (t !== null) for (let e of t) e.rv = Qn;
			K !== null && (r === null ? r = K : r.push(...K));
		}
		return e.f & 8388608 && (e.f ^= ue), d;
	} catch (t) {
		return ar(e), vn(t);
	} finally {
		e.f ^= ce, W = t, G = n, K = r, V = i, Jn = a, Ze(o), Kn = s, $n = c;
	}
}
function ar(e) {
	var t = e.deps, n = M?.is_fork;
	if (W !== null) {
		var r;
		if (n || sr(e, G), t !== null && G > 0) for (t.length = G + W.length, r = 0; r < W.length; r++) t[G + r] = W[r];
		else e.deps = t = W;
		if (Cn() && e.f & 512) for (r = G; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && G < t.length && (sr(e, G), t.length = G);
	return t;
}
function or(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = a.call(r, e);
		if (i !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[i] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (W === null || !o.call(W, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && st(c), c.ac !== null && ft(() => {
			c.ac.abort(be), c.ac = null, A(c, C);
		}), wt(c), sr(c, 0);
	}
}
function sr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) or(e, n[r]);
}
function cr(e) {
	var t = e.f;
	if (!(t & 16384)) {
		A(e, S);
		var n = U;
		U = e;
		try {
			t & 16777232 ? Fn(e) : Pn(e), Nn(e);
			var r = ir(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Zn;
		} finally {
			U = n;
		}
	}
}
function q(e) {
	var t = !!(e.f & 2);
	if (Un?.add(e), V !== null && !Kn && !(U !== null && U.f & 16384) && (Jn === null || !Jn.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < Qn && (e.rv = Qn, W === null && n !== null && n[G] === e ? G++ : W === null ? W = [e] : W.push(e));
		else {
			V.deps ??= [], o.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : o.call(r, V) || r.push(V);
		}
	}
	if (Wn && Ut.has(e)) return Ut.get(e);
	if (t) {
		var i = e;
		if (Wn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || ur(i)) && (a = St(i)), Ut.set(i, a), a;
		}
		var s = !(i.f & 512) && !Kn && V !== null && !!(V.f & 512), c = (i.f & ne) === 0;
		nr(i) && (s && (i.f |= 512), Ct(i)), s && !c && (Tt(i), lr(i));
	}
	if (N?.has(e)) return N.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function lr(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : o.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (Tt(n), lr(n));
	}
}
function ur(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ut.has(t) || t.f & 2 && ur(t)) return !0;
	return !1;
}
function dr(e) {
	var t = Kn;
	try {
		return Kn = !0, e();
	} finally {
		Kn = t;
	}
}
function fr(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (de in e) pr(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && de in n && pr(n);
		}
	}
}
function pr(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			pr(e[n], t);
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
var mr = Symbol("events"), hr = /* @__PURE__ */ new Set(), gr = /* @__PURE__ */ new Set();
function _r(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || Cr.call(t, e), !e.cancelBubble) return ft(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, at(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function vr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = _r(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && wn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function yr(e, t, n) {
	(t[mr] ??= {})[e] = n;
}
function br(e) {
	for (var t = 0; t < e.length; t++) hr.add(e[t]);
	for (var n of gr) n(e);
}
var xr = null, Sr = !1;
function Cr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	xr = e, Sr || (Sr = !0, setTimeout(() => {
		Sr = !1, xr = null;
	}));
	var o = 0, s = xr === e && e[mr];
	if (s) {
		var l = i.indexOf(s);
		if (l !== -1 && (t === document || t === window)) {
			e[mr] = t;
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
		var d = V, f = U;
		H(null), qn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[mr]?.[r];
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
			e[mr] = t, delete e.currentTarget, H(d), qn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var wr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Tr(e) {
	return wr?.createHTML(e) ?? e;
}
function Er(e) {
	var t = gn("template");
	return t.innerHTML = Tr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Dr(e, t) {
	var n = U;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function J(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (T) return Dr(E, null), E;
		i === void 0 && (i = Er(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ F(i)));
		var t = r || sn ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ F(t), s = t.lastChild;
			Dr(o, s);
		} else Dr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Or(e, t, n = "svg") {
	var r = !e.startsWith("<!>"), i = !!(t & 1), a = `<${n}>${r ? e : "<!>" + e}</${n}>`, o;
	return () => {
		if (T) return Dr(E, null), E;
		if (!o) {
			var e = /* @__PURE__ */ F(Er(a));
			if (i) for (o = document.createDocumentFragment(); /* @__PURE__ */ F(e);) o.appendChild(/* @__PURE__ */ F(e));
			else o = /* @__PURE__ */ F(e);
		}
		var t = o.cloneNode(!0);
		if (i) {
			var n = /* @__PURE__ */ F(t), r = t.lastChild;
			Dr(n, r);
		} else Dr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function kr(e, t) {
	return /* @__PURE__ */ Or(e, t, "svg");
}
function Ar() {
	if (T) return Dr(E, null), E;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = P();
	return e.append(t, n), Dr(t, n), e;
}
function Y(e, t) {
	if (T) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = E), De();
	} else e !== null && e.before(t);
}
// node_modules/svelte/src/utils.js
function jr(e) {
	return e.endsWith("capture") && e !== "gotpointercapture" && e !== "lostpointercapture";
}
var Mr = [
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
function Nr(e) {
	return Mr.includes(e);
}
var Pr = /* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split("."), Fr = {
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
function Ir(e) {
	return e = e.toLowerCase(), Fr[e] ?? e;
}
[...Pr];
var Lr = ["touchstart", "touchmove"];
function Rr(e) {
	return Lr.includes(e);
}
var zr = [
	"textarea",
	"script",
	"style",
	"title"
];
function Br(e) {
	return zr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Vr(e) {
	let t = 0, n = Gt(0), r;
	return () => {
		Cn() && (q(n), An(() => (t === 0 && (r = dr(() => e(() => en(n)))), t += 1, () => {
			at(() => {
				--t, t === 0 && (r?.(), r = void 0, en(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Hr = ie | ae;
function Ur(e, t, n, r) {
	new Wr(e, t, n, r);
}
var Wr = class {
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
	#m = Vr(() => (this.#p = Gt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = jn(() => {
			if (T) {
				let e = Ae(this.#t);
				De();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Hr), T && (this.#e = E);
	}
	#h() {
		try {
			this.#a = z(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		at(r), t && (this.#s = z(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? Te() : (t = !0, n && Ge(), this.#s !== null && Rn(this.#s, () => {
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
					yn(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (te | re)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = z(() => e(this.#e)), at(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = P(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return z(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							yn(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(M)) : this.#u === 0 && (this.#e.before(e), this.#c = null, Rn(this.#o, () => {
					this.#o = null;
				}), this.#x(M));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = z(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Hn(this.#a, e);
				let t = this.#n.pending;
				this.#o = z(() => t(this.#e));
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
		var t = U, n = V, r = k;
		qn(this.#i), H(this.#i), Ze(this.#i.ctx);
		try {
			return Ft.ensure(), e();
		} finally {
			qn(t), H(n), Ze(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && Rn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, at(() => {
			this.#d = !1, this.#p && Zt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), q(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		M?.is_fork ? (this.#a && M.skip_effect(this.#a), this.#o && M.skip_effect(this.#o), this.#s && M.skip_effect(this.#s), M.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (B(this.#a), null), this.#o &&= (B(this.#o), null), this.#s &&= (B(this.#s), null), T && (D(this.#t), Oe(), D(ke()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return z(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return yn(e, this.#i.parent), null;
				}
			}));
		};
		at(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					yn(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => yn(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Gr(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[ve] ??= e.nodeValue) && (e[ve] = n, e.nodeValue = `${n}`);
}
function Kr(e, t) {
	return Jr(e, t);
}
var qr = /* @__PURE__ */ new Map();
function Jr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: c = !0, transformError: l }) {
	un();
	var u = void 0, d = Dn(() => {
		var c = r ?? n.appendChild(P());
		Ur(c, { pending: () => {} }, (n) => {
			$e({});
			var r = k;
			if (o && (r.c = o), a && (i.$$events = a), T && Dr(n, null), u = t(n, i) || tt(), T && (U.nodes.end = E, E === null || E.nodeType !== 8 || E.data !== "]")) throw Ce(), e;
			et();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Rr(r);
					for (let e of [n, document]) {
						var a = qr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), qr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Cr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(s(hr)), gr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = qr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Cr), t.delete(e), t.size === 0 && qr.delete(r)) : t.set(e, i);
			}
			gr.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return Yr.set(u, d), u;
}
var Yr = /* @__PURE__ */ new WeakMap();
function Xr(e, t) {
	let n = Yr.get(e);
	return n ? (Yr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Zr = class {
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
			if (n) Bn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Bn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (B(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Hn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else B(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), Rn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (B(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = M, r = hn();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = P();
				i.append(a), this.#n.set(e, {
					effect: z(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, z(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else T && (this.anchor = E), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/snippet.js
function Qr(e, t, ...n) {
	var r = new Zr(e);
	jn(() => {
		let e = t() ?? null;
		r.ensure(e, e && ((t) => e(t, ...n)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function X(e, t, n = !1) {
	var r;
	T && (r = E, De());
	var i = new Zr(e), a = n ? ie : 0;
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
	jn(() => {
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
	var r = new Zr(e), i = !nt();
	jn(() => {
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
		Rn(n, () => {
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
			mn(d), d.append(u), e.items.clear();
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
		r?.has(a) ? (a.f |= se, Hn(a, document.createDocumentFragment())) : B(t[i], n);
	}
}
var ii;
function ai(e, t, n, r, a, o = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = T ? D(/* @__PURE__ */ F(u)) : u.appendChild(P());
	}
	T && De();
	var d = null, f = /* @__PURE__ */ bt(() => {
		var e = n();
		return i(e) ? e : e == null ? [] : s(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = q(f);
			_.fallback = d, si(_, n, c, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= se, li(d, null, c)) : Bn(d) : Rn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: jn(() => {
			var e = q(f), i = e.length;
			let s = !1;
			T && Ae(c) === "[!" != (i === 0) && (c = ke(), D(c), Ee(!1), s = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = M, v = hn(), y = 0; y < i; y += 1) {
				T && E.nodeType === 8 && E.data === "]" && (c = E, s = !0, Ee(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Zt(S.v, b), S.i && Zt(S.i, y), v && _.unskip_effect(S.e)) : (S = ci(l, m ? c : ii ??= P(), b, x, y, a, t, n), m || (S.e.f |= se), l.set(x, S)), u.add(x);
			}
			if (i === 0 && o && !d && (m ? d = z(() => o(c)) : (d = z(() => o(ii ??= P())), d.f |= se)), i > u.size && Ie("", "", ""), T && i > 0 && D(ke()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			s && Ee(!0), q(f);
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
		if (_.f & 8192 && (Bn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
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
	var c = o & 1 ? o & 16 ? Gt(n) : /* @__PURE__ */ qt(n, !1, !1) : null, l = o & 2 ? Gt(i) : null;
	return {
		v: c,
		i: l,
		e: z(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function li(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ dn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function ui(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-element.js
function di(e, t, n, i, a, o) {
	let s = T;
	T && De();
	var c = null;
	T && E.nodeType === 1 && (c = E, De());
	var l = T ? E : e, u = new Zr(l, !1);
	jn(() => {
		let e = t() || null;
		var o = a ? a() : n || e === "svg" ? r : void 0;
		if (e === null) u.ensure(null, null);
		else return u.ensure(e, (t) => {
			if (e) {
				var n = T;
				if (c = T ? c : gn(e, o), Dr(c, c), i) {
					var r = null;
					T && Br(e) && c.append(r = document.createComment(""));
					var a = T ? /* @__PURE__ */ F(c) : c.appendChild(P());
					T && (a === null ? Ee(!1) : D(a)), i(c, a), r?.remove();
				}
				U.nodes.end = c, n || t.before(c);
			}
			T && D(t);
		}), () => {};
	}, ie), wn(() => {}), s && (Ee(!0), D(l));
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function fi(e, t, n) {
	On(() => {
		var r = dr(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			An(() => {
				var e = n();
				fr(e), i && Me(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/client/dom/elements/attachments.js
function pi(e, t) {
	var n = void 0, r;
	Mn(() => {
		n !== (n = t()) && (r &&= (B(r), null), n && (r = z(() => {
			On(() => n(e));
		})));
	});
}
// node_modules/clsx/dist/clsx.mjs
function mi(e) {
	var t, n, r = "";
	if (typeof e == "string" || typeof e == "number") r += e;
	else if (typeof e == "object") {
		if (Array.isArray(e)) {
			var i = e.length;
			for (t = 0; t < i; t++) e[t] && (n = mi(e[t])) && (r && (r += " "), r += n);
		} else for (n in e) e[n] && (r && (r += " "), r += n);
	}
	return r;
}
function hi() {
	for (var e, t, n = 0, r = "", i = arguments.length; n < i; n++) (e = arguments[n]) && (t = mi(e)) && (r && (r += " "), r += t);
	return r;
}
// node_modules/svelte/src/internal/shared/attributes.js
function gi(e) {
	return typeof e == "object" ? hi(e) : e ?? "";
}
var _i = [..." 	\n\r\f\xA0\v﻿"];
function vi(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || _i.includes(r[o - 1])) && (s === r.length || _i.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
function yi(e, t = !1) {
	var n = t ? " !important;" : ";", r = "";
	for (var i of Object.keys(e)) {
		var a = e[i];
		a != null && a !== "" && (r += " " + i + ": " + a + n);
	}
	return r;
}
function bi(e) {
	return e[0] !== "-" || e[1] !== "-" ? e.toLowerCase() : e;
}
function xi(e, t) {
	if (t) {
		var n = "", r, i;
		if (Array.isArray(t) ? (r = t[0], i = t[1]) : r = t, e) {
			e = String(e).replaceAll(/\/\*.*?\*\//g, "").trim();
			var a = !1, o = 0, s = !1, c = [];
			r && c.push(...Object.keys(r).map(bi)), i && c.push(...Object.keys(i).map(bi));
			var l = 0, u = -1;
			let t = e.length;
			for (var d = 0; d < t; d++) {
				var f = e[d];
				if (s ? f === "/" && e[d - 1] === "*" && (s = !1) : a ? a === f && (a = !1) : f === "/" && e[d + 1] === "*" ? s = !0 : f === "\"" || f === "'" ? a = f : f === "(" ? o++ : f === ")" && o--, !s && a === !1 && o === 0) {
					if (f === ":" && u === -1) u = d;
					else if (f === ";" || d === t - 1) {
						if (u !== -1) {
							var p = bi(e.substring(l, u).trim());
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
		return r && (n += yi(r)), i && (n += yi(i, !0)), n = n.trim(), n === "" ? null : n;
	}
	return e == null ? null : String(e);
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Si(e, t, n, r, i, a) {
	var o = e[ge];
	if (T || o !== n || o === void 0) {
		var s = vi(n, r, a);
		T && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ge] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/style.js
function Ci(e, t = {}, n, r) {
	for (var i in n) {
		var a = n[i];
		t[i] !== a && (n[i] == null ? e.style.removeProperty(i) : e.style.setProperty(i, a, r));
	}
}
function wi(e, t, n, r) {
	var i = e[_e];
	if (T || i !== t) {
		var a = xi(t, r);
		T && a === e.getAttribute("style") || (a == null ? e.removeAttribute("style") : e.style.cssText = a), e[_e] = t;
	} else r && (Array.isArray(r) ? (Ci(e, n?.[0], r[0]), Ci(e, n?.[1], r[1], "important")) : Ci(e, n, r));
	return r;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function Ti(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function Ei(e, t) {
	var n = !("__defaultValue" in e);
	(n || e.__defaultValue !== t) && (e.__defaultValue = t, Di(e, !n || "__value" in e));
}
function Di(e, t) {
	var n = e.__defaultValue, r = e.multiple, a = r ? n ?? [] : null;
	if (!r || i(a)) {
		var o = e.selectedIndex, s = t && r ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = Ai(c);
			Ti(c, r ? a.includes(l) : an(l, n));
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
function Oi(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!i(t)) return we();
		for (var r of e.options) r.selected = t.includes(Ai(r));
	} else {
		for (r of e.options) if (an(Ai(r), t)) {
			r.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function ki(e) {
	var t = new MutationObserver((t) => {
		t.every(ji) || ("__defaultValue" in e && Di(e, !1), "__value" in e && Oi(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), wn(() => {
		t.disconnect();
	});
}
function Ai(e) {
	return "__value" in e ? e.__value : e.value;
}
function ji(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Mi = Symbol("class"), Ni = Symbol("style"), Pi = Symbol("is custom element"), Fi = Symbol("is html"), Ii = xe ? "link" : "LINK", Li = xe ? "input" : "INPUT", Ri = xe ? "option" : "OPTION", zi = xe ? "select" : "SELECT", Bi = xe ? "progress" : "PROGRESS";
function Vi(e) {
	if (T) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Z(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Z(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ye] = n, at(n), dt();
	}
}
function Z(e, t, n, r) {
	var i = Wi(e);
	if (!(T && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Ii))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[me] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Ki(e).has(t) ? e[t] = n : (!T || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Hi(e, n, r, i, a = !1, o = !1) {
	T && a && e.nodeName === Li && ("defaultValue" in r || "defaultChecked" in r || Vi(e));
	var s = Wi(e), c = s[Pi], l = !s[Fi];
	let u = T && c;
	u && Ee(!1);
	var d = n || {}, f = e.nodeName === Ri, p = e.nodeName === zi;
	try {
		for (var m in n) !(m in r) && m[0] + m[1] !== "$$" && (r[m] = null);
		r.class ? r.class = gi(r.class) : (i || r[Mi]) && (r.class = null), r[Ni] && (r.style ??= null);
		var h = Ki(e);
		if (e.nodeName === Li && "type" in r && ("value" in r || "__value" in r)) {
			var g = r.type;
			(g !== d.type || g === void 0 && e.hasAttribute("type")) && (d.type = g, Z(e, "type", g, o));
		}
		for (let a in r) {
			let u = r[a];
			if (f && a === "value" && u == null) e.value = e.__value = "", d[a] = u;
			else if (a === "class") Si(e, e.namespaceURI === "http://www.w3.org/1999/xhtml", u, i, n?.[Mi], r[Mi]), d[a] = u, d[Mi] = r[Mi];
			else if (a === "style") wi(e, u, n?.[Ni], r[Ni]), d[a] = u, d[Ni] = r[Ni];
			else {
				var _ = d[a];
				if (u !== _ || u === void 0 && e.hasAttribute(a)) {
					d[a] = u;
					var v = a[0] + a[1];
					if (v !== "$$") {
						if (v === "on") {
							let t = {}, n = "$$" + a, r = a.slice(2);
							var y = Nr(r);
							if (jr(r) && (r = r.slice(0, -7), t.capture = !0), !y && _) {
								if (u != null) continue;
								e.removeEventListener(r, d[n], t), d[n] = null;
							}
							if (y) yr(r, e, u), br([r]);
							else if (u != null) {
								function i(e) {
									d[a].call(this, e);
								}
								d[n] = _r(r, e, i, t);
							}
						} else if (a === "style") Z(e, a, u);
						else if (a === "autofocus") lt(e, !!u);
						else if (!c && (a === "__value" || a === "value" && u != null)) e.__value = u, (_ == null || e.value !== u || u === 0 && e.nodeName === Bi) && (e.value = u);
						else if (a === "selected" && f) Ti(e, u);
						else {
							var b = a;
							l || (b = Ir(b));
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
							} else x || (c || typeof u != "string") && h.has(b) ? (e[b] = u, b in s && (s[b] = t)) : typeof u != "function" && Z(e, b, u, o);
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
function Ui(e, t, n = [], r = [], i = [], a, o = !1, s = !1) {
	pt(i, n, r, (n) => {
		var r = void 0, i = {}, c = e.nodeName === zi, l = !1;
		if (Mn(() => {
			var u = t(...n.map(q)), d = Hi(e, r, u, a, o, s);
			if (l && c) {
				var f = e;
				"defaultValue" in u && Ei(f, u.defaultValue), "value" in u && Oi(f, u.value);
			}
			for (let e of Object.getOwnPropertySymbols(i)) u[e] || B(i[e]);
			for (let t of Object.getOwnPropertySymbols(u)) {
				var p = u[t];
				t.description !== "@attach" || r && p === r[t] || (i[t] && B(i[t]), i[t] = z(() => pi(e, () => p))), d[t] = p;
			}
			r = d;
		}), c) {
			var u = e;
			On(() => {
				var e = r;
				"defaultValue" in e && Ei(u, e.defaultValue), Oi(u, e.value, !0), ki(u);
			});
		}
		l = !0;
	});
}
function Wi(e) {
	return e[he] ??= {
		[Pi]: e.nodeName.includes("-"),
		[Fi]: e.namespaceURI === n
	};
}
var Gi = /* @__PURE__ */ new Map();
function Ki(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Gi.get(t);
	if (n) return n;
	Gi.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = u(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = p(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/this.js
function qi(e, t) {
	return e === t || e?.[de] === t;
}
function Ji(e = tt(), t, n, r) {
	var i = k.r, a = U;
	return On(() => {
		var o, s;
		return An(() => {
			o = s, s = r?.() || [], dr(() => {
				qi(n(...s), e) || (t(e, ...s), o && qi(n(...o), e) && t(null, ...o));
			});
		}), () => {
			let r = a;
			for (; r !== i && r.parent !== null && r.parent.f & 33554432;) r = r.parent;
			let o = () => {
				s && qi(n(...s), e) && t(null, ...s);
			}, c = r.teardown;
			r.teardown = () => {
				o(), c?.();
			};
		};
	}), e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Yi = !1;
function Xi(e) {
	var t = Yi;
	try {
		return Yi = !1, [e(), Yi];
	} finally {
		Yi = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
var Zi = {
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
function Qi(e, t, n) {
	return new Proxy({
		props: e,
		exclude: t
	}, Zi);
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
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ _t(r), q(u)) : (c && (c = !1, s = o ? dr(r) : r), s);
	let f;
	if (a) {
		var p = de in e || pe in e;
		f = l(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Xi(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ve(t), f(m)));
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
	a && q(y);
	var b = U;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? q(y) : i && a ? nn(e) : e;
			return Jt(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Wn && v || b.f & 16384 ? y.v : q(y);
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
		#e = /* @__PURE__ */ Kt(nn({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return q(this.#e);
		}
		set props(e) {
			Jt(this.#e, e, !0);
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
			this.app ||= Kr(t, {
				target: this,
				props: this.props
			}), i || (this.style.visibility = "hidden", e.then(() => {
				this.style.visibility = "";
			}));
		}
		disconnectedCallback() {
			this.app && Xr(this.app), this.app = null;
		}
	});
}
function ra(e, t) {
	let n = nn({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			$t(r);
		}
	}), r = /* @__PURE__ */ Kt(0);
	Tn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, $t(r);
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
	return Tn(() => {
		let n = e(), c = t();
		q(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function ia(e) {
	let t = nn({ ...e()?.params?.() ?? {} });
	return Tn(() => {
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
// plugins/desk/panel.svelte.js
var oa = 160, sa = 120, ca = 360, la = 12;
function ua(e) {
	return Math.round(Math.max(oa, Math.min(window.innerHeight - sa, e)));
}
var da = Number(localStorage.getItem("desk-console-h")), fa = window.matchMedia("(max-width: 768px)"), Q = nn({
	h: ua(da > 0 ? da : ca),
	mobile: fa.matches
});
fa.addEventListener("change", (e) => {
	Q.mobile = e.matches;
}), window.addEventListener("resize", () => {
	Q.h = ua(Q.h);
});
function pa(e) {
	Q.h = ua(e), localStorage.setItem("desk-console-h", String(Q.h));
}
function ma() {
	pa(ca);
}
function ha() {
	try {
		let e = JSON.parse(localStorage.getItem("desk-console-tabs") || "[]");
		return Array.isArray(e) ? e.filter((e) => e && typeof e.id == "string").slice(0, la) : [];
	} catch {
		return [];
	}
}
var $ = nn({ list: ha() });
function ga() {
	localStorage.setItem("desk-console-tabs", JSON.stringify($.list));
}
function _a(e) {
	if (!e.session) return;
	let t = {
		id: e.session,
		item: e.item || "",
		ro: e.ro || ""
	}, n = $.list.findIndex((e) => e.id === t.id);
	if (n < 0) $.list = [...$.list, t].slice(-12);
	else if ($.list[n].item !== t.item || $.list[n].ro !== t.ro) $.list[n] = t;
	else return;
	ga();
}
function va(e, t) {
	e.openSession(t.id, {
		item: t.item,
		readonly: t.ro === "1"
	});
}
function ya(e, t, n) {
	let r = $.list.findIndex((e) => e.id === n);
	if (r < 0 || ($.list = $.list.filter((e) => e.id !== n), ga(), t !== n)) return;
	let i = $.list[Math.min(r, $.list.length - 1)];
	i ? va(e, i) : e.go({
		session: "",
		item: "",
		ro: ""
	});
}
// node_modules/@lucide/svelte/dist/utils/defaultAttributes.js
var ba = {
	xmlns: "http://www.w3.org/2000/svg",
	width: 24,
	height: 24,
	viewBox: "0 0 24 24",
	fill: "none",
	stroke: "currentColor",
	"stroke-width": 2,
	"stroke-linecap": "round",
	"stroke-linejoin": "round"
}, xa = (...e) => e.filter((e, t, n) => !!e && e.trim() !== "" && n.indexOf(e) === t).join(" ").trim();
// node_modules/@lucide/svelte/dist/utils/buildLucideIconNode.js
function Sa(e) {
	return e != null;
}
function Ca(e, t = {}) {
	let n = t.attributeNames ?? {}, r = (e) => n[e] ?? e, i = e.size ?? e.width ?? ba.width, a = e.size ?? e.height ?? ba.height, o = e.aliases?.filter((e) => typeof e == "string" && e.trim() !== "").map((e) => `lucide-${e}`) ?? [], s = [...e.name ? [`lucide-${e.name}`] : [], ...o], c = t.className?.split(" ").filter(Boolean) ?? [], l = t.includeDefaultClasses === !1 ? xa(...c) : xa("lucide", ...s, ...c), u = t.absoluteStrokeWidth ? Number(t.strokeWidth ?? ba["stroke-width"]) * Number(e.size ?? e.width ?? ba.width) / Number(t.size ?? t.width ?? ba.width) : t.strokeWidth ?? ba["stroke-width"];
	return [
		"svg",
		{
			...Object.entries(ba).reduce((e, [t, n]) => (e[r(t)] = n, e), {}),
			..."color" in t && t.color && { [r("stroke")]: t.color },
			..."size" in t && Sa(t.size) && {
				[r("width")]: t.size,
				[r("height")]: t.size
			},
			..."width" in t && Sa(t.width) && { [r("width")]: t.width },
			..."height" in t && Sa(t.height) && { [r("height")]: t.height },
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
var wa = (e) => {
	for (let t in e) if (t.startsWith("aria-") || t === "role" || t === "title") return !0;
	return !1;
}, Ta = Symbol("lucide-context"), Ea = () => Qe(Ta), Da = /* @__PURE__ */ new Set([
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
]), Oa = /* @__PURE__ */ kr("<svg><!><!></svg>");
function ka(e, t) {
	$e(t, !0);
	let n = Ea() ?? {}, r = ta(t, "color", 19, () => n.color ?? "currentColor"), i = ta(t, "size", 19, () => n.size ?? 24), a = ta(t, "width", 19, i), o = ta(t, "height", 19, i), s = ta(t, "strokeWidth", 19, () => n.strokeWidth ?? 2), c = ta(t, "absoluteStrokeWidth", 19, () => n.absoluteStrokeWidth ?? !1), l = ta(t, "nonScalingStroke", 19, () => n.nonScalingStroke ?? !1), u = ta(t, "iconNode", 19, () => []), d = ta(t, "icon", 19, () => ({
		node: u(),
		aliases: [],
		size: 24
	})), f = /* @__PURE__ */ Qi(t, Da), p = /* @__PURE__ */ j(() => !!t.children || wa(f)), m = /* @__PURE__ */ j(() => Ca(d(), {
		color: r(),
		width: a(),
		height: o(),
		strokeWidth: s(),
		absoluteStrokeWidth: c(),
		nonScalingStroke: l(),
		className: xa("lucide-icon", n.class),
		hasA11yProp: q(p),
		attributes: f
	})), h = /* @__PURE__ */ j(() => b(q(m), 3)), _ = /* @__PURE__ */ j(() => q(h)[1]), v = /* @__PURE__ */ j(() => y(q(h)[2], () => [], !0)), x = /* @__PURE__ */ j(() => ({
		...q(_),
		class: [...q(_).class.split(" "), t.class]
	}));
	var S = Oa();
	Ui(S, () => ({ ...q(x) }));
	var C = I(S);
	ai(C, 17, () => q(v), ti, (e, t) => {
		var n = /* @__PURE__ */ j(() => b(q(t), 2));
		let r = () => q(n)[0], i = () => q(n)[1];
		var a = Ar();
		di(fn(a), r, !0, (e, t) => {
			Ui(e, () => ({ ...i() }));
		}), Y(e, a);
	}), Qr(L(C), () => t.children ?? g), O(S), Y(e, S), et();
}
// node_modules/@lucide/svelte/dist/icons/chevron-down.svelte
var Aa = {
	name: "chevron-down",
	size: 24,
	node: [["path", { d: "m6 9 6 6 6-6" }]]
};
Aa.node;
var ja = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ma(e, t) {
	let n = /* @__PURE__ */ Qi(t, ja);
	ka(e, ea(() => n, { get icon() {
		return Aa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/chevron-up.svelte
var Na = {
	name: "chevron-up",
	size: 24,
	node: [["path", { d: "m18 15-6-6-6 6" }]]
};
Na.node;
var Pa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Fa(e, t) {
	let n = /* @__PURE__ */ Qi(t, Pa);
	ka(e, ea(() => n, { get icon() {
		return Na;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/x.svelte
var Ia = {
	name: "x",
	size: 24,
	node: [["path", { d: "M18 6 6 18" }], ["path", { d: "m6 6 12 12" }]]
};
Ia.node;
var La = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ra(e, t) {
	let n = /* @__PURE__ */ Qi(t, La);
	ka(e, ea(() => n, { get icon() {
		return Ia;
	} }));
}
// plugins/desk/Console.svelte
var za = /* @__PURE__ */ J("<div class=\"grip svelte-qk9xya\" role=\"slider\" aria-orientation=\"vertical\" aria-label=\"console height\" aria-valuemin=\"160\" tabindex=\"0\"></div>"), Ba = /* @__PURE__ */ J("<div role=\"tab\" data-testid=\"console-tab\"><button type=\"button\" class=\"tl svelte-qk9xya\"> </button> <button type=\"button\" class=\"tx svelte-qk9xya\" data-testid=\"console-tab-close\"><!></button></div>"), Va = /* @__PURE__ */ J("<button type=\"button\" class=\"ic svelte-qk9xya\" aria-label=\"Minimize\" title=\"Minimize (Esc)\" data-testid=\"session-panel-close\"><!></button>"), Ha = /* @__PURE__ */ J("<button type=\"button\" class=\"ic svelte-qk9xya\" aria-label=\"Restore\" title=\"Restore\" data-testid=\"console-restore\"><!></button>"), Ua = /* @__PURE__ */ J("<div class=\"plugin-host\"></div>"), Wa = /* @__PURE__ */ J("<p class=\"faint none svelte-qk9xya\">No plugin shows sessions (slot <span class=\"mono\">session</span>).</p>"), Ga = /* @__PURE__ */ J("<div class=\"body svelte-qk9xya\"><!></div>"), Ka = /* @__PURE__ */ J("<section tabindex=\"-1\" aria-label=\"sessions console\"><!> <div class=\"bar svelte-qk9xya\"><div class=\"tabs svelte-qk9xya\" role=\"tablist\"></div> <!></div> <!></section>");
function qa(e, t) {
	$e(t, !0);
	let n = /* @__PURE__ */ j(() => t.url.session || ""), r = /* @__PURE__ */ j(() => t.url.item || ""), i = /* @__PURE__ */ j(() => t.url.ro === "1");
	Tn(() => {
		_a({
			session: q(n),
			item: q(r),
			ro: t.url.ro || ""
		});
	});
	let a = /* @__PURE__ */ j(() => t.api.contributions("session").at(-1)), o = ra(() => t.api, () => $.list.length ? "api/sessions?since=720h" : null), s = /* @__PURE__ */ j(() => Object.fromEntries((o.data ?? []).map((e) => [e.id, e]))), c = (e) => {
		let t = q(s)[e];
		return t?.ticket?.pool || t?.display_name || e;
	}, l = (e) => {
		let t = q(s)[e];
		return [
			t?.display_name || t?.ticket?.title,
			t?.model,
			e
		].filter(Boolean).join("\n");
	};
	Tn(() => {
		if (!Q.mobile || !q(n)) return;
		let e = document.body.style.overflow;
		return document.body.style.overflow = "hidden", () => {
			document.body.style.overflow = e;
		};
	});
	let u = /* @__PURE__ */ Kt(void 0);
	Tn(() => {
		q(n) && q(u)?.focus({ preventScroll: !0 });
	});
	function d(e) {
		let t = e.currentTarget;
		t.setPointerCapture(e.pointerId);
		let n = q(u).getBoundingClientRect().bottom, r = (e) => pa(n - e.clientY), i = () => {
			t.removeEventListener("pointermove", r), t.removeEventListener("pointerup", i), t.removeEventListener("pointercancel", i);
		};
		t.addEventListener("pointermove", r), t.addEventListener("pointerup", i), t.addEventListener("pointercancel", i);
	}
	function f(e) {
		e.key === "ArrowUp" ? (e.preventDefault(), pa(Q.h + 24)) : e.key === "ArrowDown" && (e.preventDefault(), pa(Q.h - 24));
	}
	var p = Ar(), m = fn(p), h = (e) => {
		var o = Ka();
		let s, p;
		var m = I(o), h = (e) => {
			var t = za();
			R(() => Z(t, "aria-valuenow", Q.h)), yr("pointerdown", t, d), yr("dblclick", t, function(...e) {
				ma?.apply(this, e);
			}), yr("keydown", t, f), Y(e, t);
		};
		X(m, (e) => {
			q(n) && !Q.mobile && e(h);
		});
		var g = L(m, 2), _ = I(g);
		ai(_, 21, () => $.list, (e) => e.id, (e, r) => {
			var i = Ba();
			let a;
			var o = I(i), s = pn(o, !0), u = L(o, 2);
			Ra(I(u), { size: 12 }), O(u), O(i), R((e, t, o) => {
				a = Si(i, 1, "tab svelte-qk9xya", null, a, { on: q(r).id === q(n) }), Z(i, "aria-selected", q(r).id === q(n)), Z(i, "data-session", q(r).id), Z(i, "title", e), Gr(s, t), Z(u, "aria-label", o);
			}, [
				() => l(q(r).id),
				() => c(q(r).id),
				() => "Close tab " + c(q(r).id)
			]), yr("click", o, () => q(r).id === q(n) && !Q.mobile ? t.api.closeSession() : va(t.api, q(r))), yr("click", u, () => ya(t.api, q(n), q(r).id)), Y(e, i);
		}), O(_);
		var v = L(_, 2), y = (e) => {
			var n = Va();
			Ma(I(n), { size: 16 }), O(n), yr("click", n, () => t.api.closeSession()), Y(e, n);
		}, b = (e) => {
			var n = Ha();
			Fa(I(n), { size: 16 }), O(n), yr("click", n, () => va(t.api, $.list.at(-1) ?? $.list[0])), Y(e, n);
		};
		X(v, (e) => {
			q(n) ? e(y) : e(b, -1);
		}), O(g);
		var x = L(g, 2), S = (e) => {
			var o = Ga(), s = I(o), c = (e) => {
				var o = Ar();
				ei(fn(o), () => q(n) + q(r) + q(i), (e) => {
					var o = Ua();
					fi(o, (e, t) => aa?.(e, t), () => ({
						api: t.api,
						slot: "session",
						target: {
							plugin: q(a).plugin,
							element: q(a).element
						},
						context: {
							session: q(n),
							item: q(r),
							readonly: q(i)
						}
					})), R(() => Z(o, "data-plugin", q(a).plugin)), Y(e, o);
				}), Y(e, o);
			}, l = (e) => {
				Y(e, Wa());
			};
			X(s, (e) => {
				q(a) ? e(c) : e(l, -1);
			}), O(o), Y(e, o);
		};
		X(x, (e) => {
			q(n) && e(S);
		}), O(o), Ji(o, (e) => Jt(u, e), () => q(u)), R(() => {
			s = Si(o, 1, "console svelte-qk9xya", null, s, { open: !!q(n) }), Z(o, "data-testid", q(n) ? "session-panel" : "console"), p = wi(o, "", p, { "--ch": Q.h + "px" });
		}), Y(e, o);
	};
	X(m, (e) => {
		(q(n) || $.list.length && !Q.mobile) && e(h);
	}), Y(e, p), et();
}
br([
	"pointerdown",
	"dblclick",
	"keydown",
	"click"
]);
// kit/format.js
function Ja(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
// plugins/desk/Drawer.svelte
var Ya = /* @__PURE__ */ J("<div class=\"err svelte-gsmuzk\"> </div>"), Xa = /* @__PURE__ */ J("<div class=\"plugin-host\"></div>"), Za = /* @__PURE__ */ J("<div class=\"row actions svelte-gsmuzk\" data-testid=\"drawer-actions\"></div>"), Qa = /* @__PURE__ */ J("<h3 class=\"svelte-gsmuzk\"> </h3>"), $a = /* @__PURE__ */ J("<!> <div data-testid=\"drawer-plugin\"><div class=\"plugin-host\"></div></div>", 1), eo = /* @__PURE__ */ J("<p class=\"faint\">No drawer plugins are enabled.</p>"), to = /* @__PURE__ */ J("<!> <!>", 1), no = /* @__PURE__ */ J("<p class=\"faint\">Loading…</p>"), ro = /* @__PURE__ */ J("<div class=\"backdrop svelte-gsmuzk\" role=\"presentation\"></div> <aside class=\"drawer svelte-gsmuzk\" data-testid=\"drawer\" aria-label=\"ticket\"><header class=\"svelte-gsmuzk\"><span class=\"mono muted id svelte-gsmuzk\"> </span> <button class=\"ghost icon\" aria-label=\"close\"><!></button></header> <div class=\"scrollbody svelte-gsmuzk\"><!></div></aside>", 1);
function io(e, t) {
	$e(t, !0);
	let n = ra(() => t.api, () => "api/ticket/" + encodeURIComponent(t.id)), r = () => t.api.go({ ticket: "" }), i = (e, t) => ({
		kind: e.kind,
		label: e.labels ?? [],
		pool: e.pool || "",
		workflow: Ja(t?.ticket ?? e).workflow || ""
	}), a = (e, n, r) => t.api.contributions(e, i(n, r)).sort((e, t) => Number(t.source === "builtin") - Number(e.source === "builtin")), o = (e) => qe({
		ticket: e.ticket,
		run: e.run,
		trace: e.trace,
		comments: e.comments,
		sessions: e.sessions
	});
	var s = ro();
	vr("keydown", on, (e) => {
		e.key === "Escape" && (e.preventDefault(), r());
	});
	var c = fn(s), l = L(c, 2), u = I(l), d = I(u), f = pn(d, !0), p = L(d, 2);
	Ra(I(p), { size: 14 }), O(p), O(u);
	var m = L(u, 2), h = I(m), g = (e) => {
		var t = Ya(), r = pn(t, !0);
		R(() => Gr(r, n.error)), Y(e, t);
	}, _ = (e) => {
		let r = /* @__PURE__ */ j(() => n.data.ticket), i = /* @__PURE__ */ j(() => o(n.data)), s = /* @__PURE__ */ j(() => a("action", q(r), n.data.run));
		var c = to(), l = fn(c), u = (e) => {
			var n = Za();
			ai(n, 21, () => q(s), (e) => e.plugin + e.element, (e, n) => {
				var r = Xa();
				fi(r, (e, t) => aa?.(e, t), () => ({
					api: t.api,
					slot: "action",
					target: {
						plugin: q(n).plugin,
						element: q(n).element
					},
					context: q(i)
				})), R(() => Z(r, "data-plugin", q(n).plugin)), Y(e, r);
			}), O(n), Y(e, n);
		};
		X(l, (e) => {
			q(s).length && e(u);
		}), ai(L(l, 2), 17, () => a("drawer", q(r), n.data.run), (e) => e.plugin + e.element, (e, n) => {
			var r = $a(), a = fn(r), o = (e) => {
				var t = Qa(), r = pn(t, !0);
				R(() => Gr(r, q(n).title)), Y(e, t);
			};
			X(a, (e) => {
				q(n).title && e(o);
			});
			var s = L(a, 2), c = I(s);
			fi(c, (e, t) => aa?.(e, t), () => ({
				api: t.api,
				slot: "drawer",
				target: {
					plugin: q(n).plugin,
					element: q(n).element
				},
				context: q(i)
			})), O(s), R(() => {
				Z(s, "data-plugin", q(n).plugin), Z(c, "data-plugin", q(n).plugin);
			}), Y(e, r);
		}, (e) => {
			Y(e, eo());
		}), Y(e, c);
	}, v = (e) => {
		Y(e, no());
	};
	X(h, (e) => {
		n.error ? e(g) : n.data ? e(_, 1) : e(v, -1);
	}), O(m), O(l), R(() => Gr(f, t.id)), yr("click", c, r), yr("click", p, r), Y(e, s), et();
}
br(["click"]);
// plugins/desk/Frame.svelte
var ao = /* @__PURE__ */ J("<div class=\"brand svelte-e6uu72\"><span class=\"mark svelte-e6uu72\"></span>Jobshop <span class=\"faint\">Desk</span></div>"), oo = /* @__PURE__ */ J("<div class=\"navhost plugin-host svelte-e6uu72\" data-testid=\"nav-host\"></div>"), so = /* @__PURE__ */ J("<div class=\"plugin-host\"></div>"), co = /* @__PURE__ */ J("<p class=\"faint pad\" data-testid=\"view-missing\">No plugin shows <span class=\"mono\"> </span>.</p>"), lo = /* @__PURE__ */ J("<div data-testid=\"desk-app\"><!> <!> <div class=\"col svelte-e6uu72\"><main class=\"svelte-e6uu72\"><!></main> <!></div> <!></div>");
function uo(e, t) {
	$e(t, !0);
	let n = ia(() => t.api), r = /* @__PURE__ */ j(() => t.api.contributions("view")), i = /* @__PURE__ */ j(() => n.view || "attention"), a = /* @__PURE__ */ j(() => q(r).find((e) => e.id === q(i))), o = /* @__PURE__ */ j(() => t.api.contributions("nav").at(-1)), s = /* @__PURE__ */ j(() => Q.mobile), c = /* @__PURE__ */ j(() => q(o) ? q(s) ? q(o).placement?.narrow || "bottom" : q(o).placement?.wide || "left" : "none"), l = /* @__PURE__ */ j(() => q(c) === "left" || q(c) === "right" || q(s) && q(c) === "bottom");
	function u(e) {
		e.key === "Escape" && n.session && !n.ticket && !e.defaultPrevented && t.api.closeSession();
	}
	var d = lo();
	vr("keydown", on, u);
	let f;
	var p = I(d), m = (e) => {
		Y(e, ao());
	};
	X(p, (e) => {
		q(l) && e(m);
	});
	var h = L(p, 2), g = (e) => {
		var n = Ar();
		ei(fn(n), () => q(o).plugin + q(o).element, (e) => {
			var n = oo();
			fi(n, (e, t) => aa?.(e, t), () => ({
				api: t.api,
				slot: "nav",
				target: {
					plugin: q(o).plugin,
					element: q(o).element
				},
				context: {
					dock: q(c),
					narrow: q(s)
				}
			})), R(() => Z(n, "data-plugin", q(o).plugin)), Y(e, n);
		}), Y(e, n);
	};
	X(h, (e) => {
		q(o) && e(g);
	});
	var _ = L(h, 2), v = I(_), y = I(v), b = (e) => {
		var n = Ar();
		ei(fn(n), () => q(a).plugin + q(a).element, (e) => {
			var n = so();
			fi(n, (e, t) => aa?.(e, t), () => ({
				api: t.api,
				slot: "view",
				target: {
					plugin: q(a).plugin,
					element: q(a).element
				}
			})), R(() => Z(n, "data-plugin", q(a).plugin)), Y(e, n);
		}), Y(e, n);
	}, x = (e) => {
		var t = co(), n = pn(L(I(t)), !0);
		Oe(), O(t), R(() => Gr(n, q(i))), Y(e, t);
	};
	X(y, (e) => {
		q(a) ? e(b) : e(x, -1);
	}), O(v), qa(L(v, 2), {
		get api() {
			return t.api;
		},
		get url() {
			return n;
		}
	}), O(_);
	var S = L(_, 2), C = (e) => {
		io(e, {
			get api() {
				return t.api;
			},
			get id() {
				return n.ticket;
			}
		});
	};
	X(S, (e) => {
		n.ticket && e(C);
	}), O(d), R(() => {
		f = Si(d, 1, "shell svelte-e6uu72", null, f, {
			narrow: q(s),
			hasbrand: q(l)
		}), Z(d, "data-dock", q(c));
	}), Y(e, d), et();
}
// plugins/desk/index.js
na("desk-app", uo);
