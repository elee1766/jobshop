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
	return D(/* @__PURE__ */ cn(E));
}
function O(t) {
	if (T) {
		if (/* @__PURE__ */ cn(E) !== null) throw Ce(), e;
		E = t;
	}
}
function Oe(e = 1) {
	if (T) {
		for (var t = e, n = E; t--;) n = /* @__PURE__ */ cn(n);
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
		var i = /* @__PURE__ */ cn(n);
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
// node_modules/svelte/src/internal/shared/context.js
function Ke(e) {
	let t = e.p;
	for (; t !== null && t.c === null;) t = t.p;
	return t?.c ?? null;
}
function qe(e, t) {
	return e === null && Pe(t), e.c ??= new Map(Ke(e) || void 0);
}
// node_modules/svelte/src/internal/client/context.js
var k = null;
function Je(e) {
	k = e;
}
function Ye(e) {
	return qe(k, "getContext").get(e);
}
function Xe(e, t = !1, n) {
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
function Ze(e) {
	var t = k, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) Sn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, k = t.p, Qe(e);
}
function Qe(e = {}) {
	return c(e, fe, { value: !0 }), e;
}
function $e() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var et = [];
function tt() {
	var e = et;
	et = [], _(e);
}
function A(e) {
	if (et.length === 0 && !Et) {
		var t = et;
		queueMicrotask(() => {
			t === et && tt();
		});
	}
	et.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var nt = ~(C | w | S);
function j(e, t) {
	e.f = e.f & nt | t;
}
function rt(e) {
	e.f & 512 || e.deps === null ? j(e, S) : j(e, w);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function it(e, t) {
	e.f & 2048 ? t.set(e, C) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, w), j(e, S);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function at(e, t) {
	if (t) {
		let t = document.body;
		e.autofocus = !0, A(() => {
			document.activeElement === t && e.focus();
		});
	}
}
var ot = !1;
function st() {
	ot || (ot = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[ye]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function ct(e) {
	var t = V, n = W;
	U(null), Un(null);
	try {
		return e();
	} finally {
		U(t), Un(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function lt(e, t, n, r) {
	let i = $e() ? pt : gt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = W, c = ut(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				hn(e, s);
			}
			dt();
		}
	}
	var d = ft();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ ht(e))).then(u).catch((e) => hn(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), dt());
	}) : f();
}
function ut() {
	var e = W, t = V, n = k, r = N;
	return function(i = !0) {
		Un(e), U(t), Je(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function dt(e = !0) {
	Un(null), U(null), Je(null), e && N?.deactivate();
}
function ft() {
	var e = W, t = e.b, n = N, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function pt(e) {
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
var mt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function ht(e, n, r) {
	let i = W;
	i === null && Fe();
	var a = void 0, o = Ht(t), s = !V, c = /* @__PURE__ */ new Set();
	return Tn(() => {
		var t = W, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== be && n.reject(e);
			}).finally(dt);
		} catch (e) {
			n.reject(e), dt();
		}
		var r = N;
		if (s) {
			if (t.f & 32768) var l = ft();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(mt);
			else for (let e of c.values()) e.reject(mt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== mt && (r.activate(), t ? (o.f |= ue, Jt(o, t)) : (o.f & 8388608 && (o.f ^= ue), Jt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), bn(() => {
		for (let e of c) e.reject(mt);
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
function M(e) {
	let t = /* @__PURE__ */ pt(e);
	return Gn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function gt(e) {
	let t = /* @__PURE__ */ pt(e);
	return t.equals = Ne, t;
}
function _t(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) B(t[n]);
	}
}
function vt(e) {
	var n, r = W, i = e.parent;
	if (!Vn && i !== null && e.v !== t && i.f & 24576) return Se(), e.v;
	Un(i);
	try {
		_t(e), n = er(e);
	} finally {
		Un(r);
	}
	return n;
}
function yt(e) {
	var t = vt(e);
	!e.equals(t) && (e.wv = Zn(), (!N?.is_fork || e.deps === null) && (N === null ? e.v = t : (N.capture(e, t, !0), Ct?.capture(e, t, !0)), e.deps === null)) ? j(e, S) : Vn || (wt === null ? rt(e) : (yn() || N?.is_fork) && wt.set(e, t));
}
function bt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && ct(() => {
		t.ac.abort(be), t.ac = null;
	}), t.fn !== null && (t.teardown = g), rr(t, 0), jn(t));
}
function xt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && ir(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var St = null, N = null, Ct = null, wt = null, Tt = null, Et = !1, Dt = !1, Ot = null, kt = null, At = 0, jt = 1, Mt = class e {
	id = jt++;
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
		St === null ? St = this : (St.next = this, this.prev = St), St = this;
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
			for (var r of n.d) j(r, C), t(r);
			for (r of n.m) j(r, w), t(r);
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
		for (let [e, t] of this.#s) j(e, t), this.schedule(e);
		this.apply();
		for (var t = Ot = [], n = [], r = kt = []; this.#a.length > 0;) {
			At++ > 1e3 && (this.#b(), Nt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Rt(e), this.#d() || this.discard(), t;
			}
		}
		if (N = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (Ot = null, kt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Lt(e, t);
			r.length > 0 && N.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), Ct = this, Ft(n), Ft(t), Ct = null, this.#i?.resolve();
			var o = N;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Bt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= S;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= S : i & 4 ? t.push(r) : Qn(r) && (i & 16 && this.add_dirty_reaction(r, w), ir(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), j(i, C), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), N = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) it(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), wt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		N = this;
	}
	deactivate() {
		N = null, wt = null;
	}
	flush() {
		try {
			Dt = !0, N = this, this.#m();
		} finally {
			At = 0, Tt = null, Ot = null, kt = null, Dt = !1, N = null, wt = null, Bt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(mt);
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
		this.#u || (this.#u = !0, A(() => {
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
		if (N === null) {
			let t = N = new e();
			!Dt && A(() => {
				t.started || t.flush();
			});
		}
		return N;
	}
	apply() {
		wt = null;
	}
	schedule(e) {
		Tt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? St = e : t.prev = e, this.linked = !1;
		}
	}
};
function Nt() {
	try {
		Be();
	} catch (e) {
		hn(e, Tt);
	}
}
var Pt = null;
function Ft(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Qn(r) && (Pt = /* @__PURE__ */ new Set(), ir(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && Pn(r), Pt?.size > 0)) {
				Bt.clear();
				for (let e of Pt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Pt.has(n) && (Pt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || ir(n);
					}
				}
				Pt.clear();
			}
		}
		Pt = null;
	}
}
function It(e) {
	N.schedule(e);
}
function Lt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), j(e, S);
		for (var n = e.first; n !== null;) Lt(n, t), n = n.next;
	}
}
function Rt(e) {
	j(e, S);
	for (var t = e.first; t !== null;) Rt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var zt = /* @__PURE__ */ new Set(), Bt = /* @__PURE__ */ new Map(), Vt = !1;
function Ht(e, t) {
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
function Ut(e, t) {
	let n = Ht(e, t);
	return Gn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Wt(e, t = !1, n = !0) {
	let r = Ht(e);
	return t || (r.equals = Ne), r;
}
function Gt(e, t, n = !1) {
	return V !== null && (!H || V.f & 131072) && $e() && V.f & 4325394 && (Wn === null || !Wn.has(e)) && We(), Jt(e, n ? $t(t) : t, kt);
}
var Kt = null, qt = 0;
function Jt(e, t, n = null) {
	if (!e.equals(t)) {
		Vn ? Bt.set(e, t) : Bt.has(e) || Bt.set(e, e.v);
		var r = Mt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && vt(t), wt === null && rt(t);
		}
		e.wv = Zn(), Kt = null, qt = 0, Qt(e, C, n), Kt = null, $e() && W !== null && W.f & 1024 && !(W.f & 96) && (q === null ? Kn([e]) : q.push(e)), !r.is_fork && zt.size > 0 && !Vt && Yt();
	}
	return t;
}
function Yt() {
	Vt = !1;
	for (let e of zt) {
		e.f & 1024 && j(e, w);
		let t;
		try {
			t = Qn(e);
		} catch {
			t = !0;
		}
		t && ir(e);
	}
	zt.clear();
}
function Xt(e, t = 1) {
	var n = J(e), r = t === 1 ? n++ : n--;
	return Gt(e, n), r;
}
function Zt(e) {
	Gt(e, e.v + 1);
}
function Qt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = $e(), a = r.length;
		if (qt += a, qt > 1e5 && Kt === null && (Kt = /* @__PURE__ */ new Set()), Kt !== null) {
			if (Kt.has(e)) return;
			Kt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & C) === 0;
				if (l && j(s, t), c & 131072) zt.add(s);
				else if (c & 2) {
					var u = s;
					wt?.delete(u), Qt(u, w, n);
				} else if (l) {
					var d = s;
					c & 16 && Pt !== null && Pt.add(d), n === null ? It(d) : n.push(d);
				}
			}
		}
	}
}
function $t(e) {
	if (typeof e != "object" || !e || de in e || fe in e) return e;
	let n = p(e);
	if (n !== d && n !== f) return e;
	var r = /* @__PURE__ */ new Map(), a = i(e), o = /* @__PURE__ */ Ut(0), s = null, c = Yn, u = (e) => {
		if (Yn === c) return e();
		var t = V, n = Yn;
		U(null), Xn(c);
		var r = e();
		return U(t), Xn(n), r;
	};
	return a && r.set("length", /* @__PURE__ */ Ut(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || He();
			var i = r.get(t);
			return i === void 0 ? u(() => {
				var e = /* @__PURE__ */ Ut(n.value, s);
				return r.set(t, e), e;
			}) : Gt(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = u(() => /* @__PURE__ */ Ut(t, s));
					r.set(n, e), Zt(o);
				}
			} else Gt(i, t), Zt(o);
			return !0;
		},
		get(n, i, a) {
			if (i === de) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || l(n, i)?.writable) && (o = u(() => /* @__PURE__ */ Ut($t(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
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
			return (i !== void 0 || (W !== null || V !== null) && (!a || l(e, n)?.writable)) && (i === void 0 && (i = u(() => /* @__PURE__ */ Ut(a ? $t(e[n]) : t, s)), r.set(n, i)), J(i) === t) ? !1 : a;
		},
		set(e, n, i, c) {
			var d = r.get(n), f = n in e;
			if (a && n === "length") for (var p = i; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = u(() => /* @__PURE__ */ Ut(t, s)), r.set(p + "", m)) : Gt(m, t);
			}
			if (d === void 0) (!f || l(e, n)?.writable) && (d = u(() => /* @__PURE__ */ Ut(void 0, s)), Gt(d, $t(i)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = u(() => $t(i));
				Gt(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, i), !f) {
				if (a && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && Gt(_, v + 1);
				}
				Zt(o);
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
function en(e) {
	try {
		if (typeof e == "object" && e && de in e) return e[de];
	} catch {}
	return e;
}
function tn(e, t) {
	return Object.is(en(e), en(t));
}
var nn, rn, an, on;
function sn() {
	if (nn === void 0) {
		nn = window, rn = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		an = l(t, "firstChild").get, on = l(t, "nextSibling").get, m(e) && (e[ge] = void 0, e[he] = null, e[_e] = void 0, e.__e = void 0), m(n) && (n[ve] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function F(e) {
	return an.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function cn(e) {
	return on.call(e);
}
function I(e, t) {
	if (!T) return /* @__PURE__ */ F(e);
	var n = /* @__PURE__ */ F(E);
	if (n === null) n = E.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), D(r), r;
	}
	return t && pn(n), D(n), n;
}
function ln(e, t = !1) {
	if (!T) {
		var n = /* @__PURE__ */ F(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ cn(n) : n;
	}
	if (t) {
		if (E?.nodeType !== 3) {
			var r = P();
			return E?.before(r), D(r), r;
		}
		pn(E);
	}
	return E;
}
function L(e, t = !1) {
	if (!T) return /* @__PURE__ */ F(e);
	var n = I(e, t);
	return O(e), n;
}
function R(e, t = 1, n = !1) {
	let r = T ? E : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ cn(r);
	if (!T) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), D(a), a;
		}
		pn(r);
	}
	return D(r), r;
}
function un(e) {
	e.textContent = "";
}
function dn() {
	return !1;
}
function fn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function pn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function mn(e) {
	var t = W;
	if (t === null) return V.f |= ue, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	hn(e, t);
}
function hn(t, n) {
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
function gn(e) {
	W === null && (V === null && ze(e), Re()), Vn && Le(e);
}
function _n(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function vn(e, t) {
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
	N?.register_created_effect(r);
	var i = r;
	if (e & 4) Ot === null ? Mt.ensure().schedule(r) : Ot.push(r);
	else if (t !== null) {
		try {
			ir(r);
		} catch (e) {
			throw B(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ie));
	}
	if (i !== null && (i.parent = n, n !== null && _n(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function yn() {
	return V !== null && !H;
}
function bn(e) {
	let t = vn(8, null);
	return j(t, S), t.teardown = e, t;
}
function xn(e) {
	gn("$effect");
	var t = W.f;
	if (!V && t & 32 && k !== null && !k.i) {
		var n = k;
		(n.e ??= []).push(e);
	} else return Sn(e);
}
function Sn(e) {
	return vn(4 | oe, e);
}
function Cn(e) {
	Mt.ensure();
	let t = vn(64 | ae, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? Fn(t, () => {
			B(t), n(void 0);
		}) : (B(t), n(void 0));
	});
}
function wn(e) {
	return vn(4, e);
}
function Tn(e) {
	return vn(le | ae, e);
}
function En(e, t = 0) {
	return vn(8 | t, e);
}
function Dn(e, t = [], n = [], r = []) {
	lt(r, t, n, (t) => {
		vn(8, () => {
			e(...t.map(J));
		});
	});
}
function On(e, t = 0) {
	return vn(16 | t, e);
}
function kn(e, t = 0) {
	return vn(x | t, e);
}
function z(e) {
	return vn(32 | ae, e);
}
function An(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Vn, r = V;
		Hn(!0), U(null);
		try {
			t.call(null);
		} catch (t) {
			hn(t, e.parent);
		} finally {
			Hn(n), U(r);
		}
	}
}
function jn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && ct(() => {
			e.abort(be);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : B(n, t), n = r;
	}
}
function Mn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || B(t), t = n;
	}
}
function B(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Nn(e.nodes.start, e.nodes.end), n = !0), e.f |= re, jn(e, t && !n), rr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	An(e), e.f ^= re, e.f |= te;
	var i = e.parent;
	i !== null && i.first !== null && Pn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Nn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ cn(e);
		e.remove(), e = n;
	}
}
function Pn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function Fn(e, t, n = !0) {
	var r = [];
	e.f |= 256, In(e, r, !0);
	var i = () => {
		n && B(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function In(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ee;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				In(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Ln(e) {
	e.f &= -257, Rn(e, !0);
}
function Rn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ee, e.f & 1024 || (j(e, C), Mt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Rn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function zn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ cn(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Bn = null, Vn = !1;
function Hn(e) {
	Vn = e;
}
var V = null, H = !1;
function U(e) {
	V = e;
}
var W = null;
function Un(e) {
	W = e;
}
var Wn = null;
function Gn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (Wn ??= /* @__PURE__ */ new Set()).add(e);
}
var G = null, K = 0, q = null;
function Kn(e) {
	q = e;
}
var qn = 1, Jn = 0, Yn = Jn;
function Xn(e) {
	Yn = e;
}
function Zn() {
	return ++qn;
}
function Qn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Qn(a) && yt(a), a.wv > e.wv) return !0;
		}
		t & 512 && wt === null && j(e, S);
	}
	return !1;
}
function $n(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Wn !== null && Wn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? $n(a, t, !1) : t === a && (n ? j(a, C) : a.f & 1024 && j(a, w), It(a));
	}
}
function er(e) {
	var t = G, n = K, r = q, i = V, a = Wn, o = k, s = H, c = Yn, l = e.f;
	G = null, K = 0, q = null, V = l & 96 ? null : e, Wn = null, Je(e.ctx), H = !1, Yn = ++Jn, e.ac !== null && (ct(() => {
		e.ac.abort(be);
	}), e.ac = null);
	try {
		e.f |= ce;
		var u = e.fn, d = u();
		e.f |= ne;
		var f = tr(e);
		if ($e() && q !== null && !H && f !== null && !(e.f & 6146)) for (var p = 0; p < q.length; p++) $n(q[p], e);
		if (i !== null && i !== e) {
			if (Jn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Jn;
			if (t !== null) for (let e of t) e.rv = Jn;
			q !== null && (r === null ? r = q : r.push(...q));
		}
		return e.f & 8388608 && (e.f ^= ue), d;
	} catch (t) {
		return tr(e), mn(t);
	} finally {
		e.f ^= ce, G = t, K = n, q = r, V = i, Wn = a, Je(o), H = s, Yn = c;
	}
}
function tr(e) {
	var t = e.deps, n = N?.is_fork;
	if (G !== null) {
		var r;
		if (n || rr(e, K), t !== null && K > 0) for (t.length = K + G.length, r = 0; r < G.length; r++) t[K + r] = G[r];
		else e.deps = t = G;
		if (yn() && e.f & 512) for (r = K; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && K < t.length && (rr(e, K), t.length = K);
	return t;
}
function nr(e, n) {
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
		c.f & 512 && (c.f ^= 512), c.v !== t && rt(c), c.ac !== null && ct(() => {
			c.ac.abort(be), c.ac = null, j(c, C);
		}), bt(c), rr(c, 0);
	}
}
function rr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) nr(e, n[r]);
}
function ir(e) {
	var t = e.f;
	if (!(t & 16384)) {
		j(e, S);
		var n = W;
		W = e;
		try {
			t & 16777232 ? Mn(e) : jn(e), An(e);
			var r = er(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = qn;
		} finally {
			W = n;
		}
	}
}
function J(e) {
	var t = !!(e.f & 2);
	if (Bn?.add(e), V !== null && !H && !(W !== null && W.f & 16384) && (Wn === null || !Wn.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < Jn && (e.rv = Jn, G === null && n !== null && n[K] === e ? K++ : G === null ? G = [e] : G.push(e));
		else {
			V.deps ??= [], o.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : o.call(r, V) || r.push(V);
		}
	}
	if (Vn && Bt.has(e)) return Bt.get(e);
	if (t) {
		var i = e;
		if (Vn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || or(i)) && (a = vt(i)), Bt.set(i, a), a;
		}
		var s = !(i.f & 512) && !H && V !== null && !!(V.f & 512), c = (i.f & ne) === 0;
		Qn(i) && (s && (i.f |= 512), yt(i)), s && !c && (xt(i), ar(i));
	}
	if (wt?.has(e)) return wt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function ar(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : o.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (xt(n), ar(n));
	}
}
function or(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Bt.has(t) || t.f & 2 && or(t)) return !0;
	return !1;
}
function sr(e) {
	var t = H;
	try {
		return H = !0, e();
	} finally {
		H = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var cr = Symbol("events"), lr = /* @__PURE__ */ new Set(), ur = /* @__PURE__ */ new Set();
function dr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || gr.call(t, e), !e.cancelBubble) return ct(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, A(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function fr(e, t, n) {
	(t[cr] ??= {})[e] = n;
}
function pr(e) {
	for (var t = 0; t < e.length; t++) lr.add(e[t]);
	for (var n of ur) n(e);
}
var mr = null, hr = !1;
function gr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	mr = e, hr || (hr = !0, setTimeout(() => {
		hr = !1, mr = null;
	}));
	var o = 0, s = mr === e && e[cr];
	if (s) {
		var l = i.indexOf(s);
		if (l !== -1 && (t === document || t === window)) {
			e[cr] = t;
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
		var d = V, f = W;
		U(null), Un(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[cr]?.[r];
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
			e[cr] = t, delete e.currentTarget, U(d), Un(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var _r = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function vr(e) {
	return _r?.createHTML(e) ?? e;
}
function yr(e) {
	var t = fn("template");
	return t.innerHTML = vr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function br(e, t) {
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
		if (T) return br(E, null), E;
		i === void 0 && (i = yr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ F(i)));
		var t = r || rn ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ F(t), s = t.lastChild;
			br(o, s);
		} else br(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function xr(e, t, n = "svg") {
	var r = !e.startsWith("<!>"), i = !!(t & 1), a = `<${n}>${r ? e : "<!>" + e}</${n}>`, o;
	return () => {
		if (T) return br(E, null), E;
		if (!o) {
			var e = /* @__PURE__ */ F(yr(a));
			if (i) for (o = document.createDocumentFragment(); /* @__PURE__ */ F(e);) o.appendChild(/* @__PURE__ */ F(e));
			else o = /* @__PURE__ */ F(e);
		}
		var t = o.cloneNode(!0);
		if (i) {
			var n = /* @__PURE__ */ F(t), r = t.lastChild;
			br(n, r);
		} else br(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Sr(e, t) {
	return /* @__PURE__ */ xr(e, t, "svg");
}
function Cr() {
	if (T) return br(E, null), E;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = P();
	return e.append(t, n), br(t, n), e;
}
function X(e, t) {
	if (T) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = E), De();
	} else e !== null && e.before(t);
}
// node_modules/svelte/src/utils.js
function wr(e) {
	return e.endsWith("capture") && e !== "gotpointercapture" && e !== "lostpointercapture";
}
var Tr = [
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
function Er(e) {
	return Tr.includes(e);
}
var Dr = /* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split("."), Or = {
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
function kr(e) {
	return e = e.toLowerCase(), Or[e] ?? e;
}
[...Dr];
var Ar = ["touchstart", "touchmove"];
function jr(e) {
	return Ar.includes(e);
}
var Mr = [
	"textarea",
	"script",
	"style",
	"title"
];
function Nr(e) {
	return Mr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Pr(e) {
	let t = 0, n = Ht(0), r;
	return () => {
		yn() && (J(n), En(() => (t === 0 && (r = sr(() => e(() => Zt(n)))), t += 1, () => {
			A(() => {
				--t, t === 0 && (r?.(), r = void 0, Zt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Fr = ie | ae;
function Ir(e, t, n, r) {
	new Lr(e, t, n, r);
}
var Lr = class {
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
	#m = Pr(() => (this.#p = Ht(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = On(() => {
			if (T) {
				let e = Ae(this.#t);
				De();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Fr), T && (this.#e = E);
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
		A(r), t && (this.#s = z(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? Te() : (t = !0, n && Ge(), this.#s !== null && Fn(this.#s, () => {
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
					hn(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (te | re)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = z(() => e(this.#e)), A(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = P(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return z(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							hn(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(N)) : this.#u === 0 && (this.#e.before(e), this.#c = null, Fn(this.#o, () => {
					this.#o = null;
				}), this.#x(N));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = z(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				zn(this.#a, e);
				let t = this.#n.pending;
				this.#o = z(() => t(this.#e));
			} else this.#x(N);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		it(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = W, n = V, r = k;
		Un(this.#i), U(this.#i), Je(this.#i.ctx);
		try {
			return Mt.ensure(), e();
		} finally {
			Un(t), U(n), Je(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && Fn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, A(() => {
			this.#d = !1, this.#p && Jt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), J(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		N?.is_fork ? (this.#a && N.skip_effect(this.#a), this.#o && N.skip_effect(this.#o), this.#s && N.skip_effect(this.#s), N.oncommit(() => {
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
						var r = W;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return hn(e, this.#i.parent), null;
				}
			}));
		};
		A(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					hn(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => hn(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Rr(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[ve] ??= e.nodeValue) && (e[ve] = n, e.nodeValue = `${n}`);
}
function zr(e, t) {
	return Vr(e, t);
}
var Br = /* @__PURE__ */ new Map();
function Vr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: c = !0, transformError: l }) {
	sn();
	var u = void 0, d = Cn(() => {
		var c = r ?? n.appendChild(P());
		Ir(c, { pending: () => {} }, (n) => {
			Xe({});
			var r = k;
			if (o && (r.c = o), a && (i.$$events = a), T && br(n, null), u = t(n, i) || Qe(), T && (W.nodes.end = E, E === null || E.nodeType !== 8 || E.data !== "]")) throw Ce(), e;
			Ze();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = jr(r);
					for (let e of [n, document]) {
						var a = Br.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Br.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, gr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(s(lr)), ur.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Br.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, gr), t.delete(e), t.size === 0 && Br.delete(r)) : t.set(e, i);
			}
			ur.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return Hr.set(u, d), u;
}
var Hr = /* @__PURE__ */ new WeakMap();
function Ur(e, t) {
	let n = Hr.get(e);
	return n ? (Hr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Wr = class {
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
			if (n) Ln(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Ln(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
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
						zn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else B(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), Fn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (B(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = N, r = dn();
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
function Gr(e, t, ...n) {
	var r = new Wr(e);
	On(() => {
		let e = t() ?? null;
		r.ensure(e, e && ((t) => e(t, ...n)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function Kr(e, t, n = !1) {
	var r;
	T && (r = E, De());
	var i = new Wr(e), a = n ? ie : 0;
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
	On(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function qr(e, t) {
	return t;
}
function Jr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, c = 0; c < i; c++) {
		let n = t[c];
		Fn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Yr(e, s(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			un(d), d.append(u), e.items.clear();
		}
		Yr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Yr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= se, zn(a, document.createDocumentFragment())) : B(t[i], n);
	}
}
var Xr;
function Zr(e, t, n, r, a, o = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = T ? D(/* @__PURE__ */ F(u)) : u.appendChild(P());
	}
	T && De();
	var d = null, f = /* @__PURE__ */ gt(() => {
		var e = n();
		return i(e) ? e : e == null ? [] : s(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = J(f);
			_.fallback = d, $r(_, n, c, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= se, ti(d, null, c)) : Ln(d) : Fn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: On(() => {
			var e = J(f), i = e.length;
			let s = !1;
			T && Ae(c) === "[!" != (i === 0) && (c = ke(), D(c), Ee(!1), s = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = N, v = dn(), y = 0; y < i; y += 1) {
				T && E.nodeType === 8 && E.data === "]" && (c = E, s = !0, Ee(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Jt(S.v, b), S.i && Jt(S.i, y), v && _.unskip_effect(S.e)) : (S = ei(l, m ? c : Xr ??= P(), b, x, y, a, t, n), m || (S.e.f |= se), l.set(x, S)), u.add(x);
			}
			if (i === 0 && o && !d && (m ? d = z(() => o(c)) : (d = z(() => o(Xr ??= P())), d.f |= se)), i > u.size && Ie("", "", ""), T && i > 0 && D(ke()), !m) {
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
function Qr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function $r(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, c = e.items, l = Qr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (Ln(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= se, _ === l) ti(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), ni(e, d, _), ni(e, _, y), ti(_, y, n), d = _, p = [], m = [], l = Qr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) ti(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					ni(e, S.prev, C.next), ni(e, d, S), ni(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), ti(_, l, n), ni(e, _.prev, _.next), ni(e, _, d === null ? e.effect.first : d.next), ni(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Qr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Qr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Yr(e, s(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = Qr(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			Jr(e, w, te);
		}
	}
	a && A(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function ei(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Ht(n) : /* @__PURE__ */ Wt(n, !1, !1) : null, l = o & 2 ? Ht(i) : null;
	return {
		v: c,
		i: l,
		e: z(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function ti(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ cn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function ni(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-component.js
function ri(e, t, n) {
	var r;
	T && (r = E, De());
	var i = new Wr(e);
	On(() => {
		var e = t() ?? null;
		if (T && Ae(r) === "[" != (e !== null)) {
			var a = ke();
			D(a), i.anchor = a, Ee(!1), i.ensure(e, e && ((t) => n(t, e))), Ee(!0);
		} else i.ensure(e, e && ((t) => n(t, e)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-element.js
function ii(e, t, n, i, a, o) {
	let s = T;
	T && De();
	var c = null;
	T && E.nodeType === 1 && (c = E, De());
	var l = T ? E : e, u = new Wr(l, !1);
	On(() => {
		let e = t() || null;
		var o = a ? a() : n || e === "svg" ? r : void 0;
		if (e === null) u.ensure(null, null);
		else return u.ensure(e, (t) => {
			if (e) {
				var n = T;
				if (c = T ? c : fn(e, o), br(c, c), i) {
					var r = null;
					T && Nr(e) && c.append(r = document.createComment(""));
					var a = T ? /* @__PURE__ */ F(c) : c.appendChild(P());
					T && (a === null ? Ee(!1) : D(a)), i(c, a), r?.remove();
				}
				W.nodes.end = c, n || t.before(c);
			}
			T && D(t);
		}), () => {};
	}, ie), bn(() => {}), s && (Ee(!0), D(l));
}
// node_modules/svelte/src/internal/client/dom/elements/attachments.js
function ai(e, t) {
	var n = void 0, r;
	kn(() => {
		n !== (n = t()) && (r &&= (B(r), null), n && (r = z(() => {
			wn(() => n(e));
		})));
	});
}
// node_modules/clsx/dist/clsx.mjs
function oi(e) {
	var t, n, r = "";
	if (typeof e == "string" || typeof e == "number") r += e;
	else if (typeof e == "object") {
		if (Array.isArray(e)) {
			var i = e.length;
			for (t = 0; t < i; t++) e[t] && (n = oi(e[t])) && (r && (r += " "), r += n);
		} else for (n in e) e[n] && (r && (r += " "), r += n);
	}
	return r;
}
function si() {
	for (var e, t, n = 0, r = "", i = arguments.length; n < i; n++) (e = arguments[n]) && (t = oi(e)) && (r && (r += " "), r += t);
	return r;
}
// node_modules/svelte/src/internal/shared/attributes.js
function ci(e) {
	return typeof e == "object" ? si(e) : e ?? "";
}
var li = [..." 	\n\r\f\xA0\v﻿"];
function ui(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || li.includes(r[o - 1])) && (s === r.length || li.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
function di(e, t = !1) {
	var n = t ? " !important;" : ";", r = "";
	for (var i of Object.keys(e)) {
		var a = e[i];
		a != null && a !== "" && (r += " " + i + ": " + a + n);
	}
	return r;
}
function fi(e) {
	return e[0] !== "-" || e[1] !== "-" ? e.toLowerCase() : e;
}
function pi(e, t) {
	if (t) {
		var n = "", r, i;
		if (Array.isArray(t) ? (r = t[0], i = t[1]) : r = t, e) {
			e = String(e).replaceAll(/\/\*.*?\*\//g, "").trim();
			var a = !1, o = 0, s = !1, c = [];
			r && c.push(...Object.keys(r).map(fi)), i && c.push(...Object.keys(i).map(fi));
			var l = 0, u = -1;
			let t = e.length;
			for (var d = 0; d < t; d++) {
				var f = e[d];
				if (s ? f === "/" && e[d - 1] === "*" && (s = !1) : a ? a === f && (a = !1) : f === "/" && e[d + 1] === "*" ? s = !0 : f === "\"" || f === "'" ? a = f : f === "(" ? o++ : f === ")" && o--, !s && a === !1 && o === 0) {
					if (f === ":" && u === -1) u = d;
					else if (f === ";" || d === t - 1) {
						if (u !== -1) {
							var p = fi(e.substring(l, u).trim());
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
		return r && (n += di(r)), i && (n += di(i, !0)), n = n.trim(), n === "" ? null : n;
	}
	return e == null ? null : String(e);
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function mi(e, t, n, r, i, a) {
	var o = e[ge];
	if (T || o !== n || o === void 0) {
		var s = ui(n, r, a);
		T && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ge] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/style.js
function hi(e, t = {}, n, r) {
	for (var i in n) {
		var a = n[i];
		t[i] !== a && (n[i] == null ? e.style.removeProperty(i) : e.style.setProperty(i, a, r));
	}
}
function gi(e, t, n, r) {
	var i = e[_e];
	if (T || i !== t) {
		var a = pi(t, r);
		T && a === e.getAttribute("style") || (a == null ? e.removeAttribute("style") : e.style.cssText = a), e[_e] = t;
	} else r && (Array.isArray(r) ? (hi(e, n?.[0], r[0]), hi(e, n?.[1], r[1], "important")) : hi(e, n, r));
	return r;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function _i(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function vi(e, t) {
	var n = !("__defaultValue" in e);
	(n || e.__defaultValue !== t) && (e.__defaultValue = t, yi(e, !n || "__value" in e));
}
function yi(e, t) {
	var n = e.__defaultValue, r = e.multiple, a = r ? n ?? [] : null;
	if (!r || i(a)) {
		var o = e.selectedIndex, s = t && r ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = Si(c);
			_i(c, r ? a.includes(l) : tn(l, n));
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
function bi(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!i(t)) return we();
		for (var r of e.options) r.selected = t.includes(Si(r));
	} else {
		for (r of e.options) if (tn(Si(r), t)) {
			r.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function xi(e) {
	var t = new MutationObserver((t) => {
		t.every(Ci) || ("__defaultValue" in e && yi(e, !1), "__value" in e && bi(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), bn(() => {
		t.disconnect();
	});
}
function Si(e) {
	return "__value" in e ? e.__value : e.value;
}
function Ci(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var wi = Symbol("class"), Ti = Symbol("style"), Ei = Symbol("is custom element"), Di = Symbol("is html"), Oi = xe ? "link" : "LINK", ki = xe ? "input" : "INPUT", Ai = xe ? "option" : "OPTION", ji = xe ? "select" : "SELECT", Mi = xe ? "progress" : "PROGRESS";
function Ni(e) {
	if (T) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Pi(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Pi(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ye] = n, A(n), st();
	}
}
function Pi(e, t, n, r) {
	var i = Li(e);
	if (!(T && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Oi))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[me] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && zi(e).has(t) ? e[t] = n : (!T || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Fi(e, n, r, i, a = !1, o = !1) {
	T && a && e.nodeName === ki && ("defaultValue" in r || "defaultChecked" in r || Ni(e));
	var s = Li(e), c = s[Ei], l = !s[Di];
	let u = T && c;
	u && Ee(!1);
	var d = n || {}, f = e.nodeName === Ai, p = e.nodeName === ji;
	try {
		for (var m in n) !(m in r) && m[0] + m[1] !== "$$" && (r[m] = null);
		r.class ? r.class = ci(r.class) : (i || r[wi]) && (r.class = null), r[Ti] && (r.style ??= null);
		var h = zi(e);
		if (e.nodeName === ki && "type" in r && ("value" in r || "__value" in r)) {
			var g = r.type;
			(g !== d.type || g === void 0 && e.hasAttribute("type")) && (d.type = g, Pi(e, "type", g, o));
		}
		for (let a in r) {
			let u = r[a];
			if (f && a === "value" && u == null) e.value = e.__value = "", d[a] = u;
			else if (a === "class") mi(e, e.namespaceURI === "http://www.w3.org/1999/xhtml", u, i, n?.[wi], r[wi]), d[a] = u, d[wi] = r[wi];
			else if (a === "style") gi(e, u, n?.[Ti], r[Ti]), d[a] = u, d[Ti] = r[Ti];
			else {
				var _ = d[a];
				if (u !== _ || u === void 0 && e.hasAttribute(a)) {
					d[a] = u;
					var v = a[0] + a[1];
					if (v !== "$$") {
						if (v === "on") {
							let t = {}, n = "$$" + a, r = a.slice(2);
							var y = Er(r);
							if (wr(r) && (r = r.slice(0, -7), t.capture = !0), !y && _) {
								if (u != null) continue;
								e.removeEventListener(r, d[n], t), d[n] = null;
							}
							if (y) fr(r, e, u), pr([r]);
							else if (u != null) {
								function i(e) {
									d[a].call(this, e);
								}
								d[n] = dr(r, e, i, t);
							}
						} else if (a === "style") Pi(e, a, u);
						else if (a === "autofocus") at(e, !!u);
						else if (!c && (a === "__value" || a === "value" && u != null)) e.__value = u, (_ == null || e.value !== u || u === 0 && e.nodeName === Mi) && (e.value = u);
						else if (a === "selected" && f) _i(e, u);
						else {
							var b = a;
							l || (b = kr(b));
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
							} else x || (c || typeof u != "string") && h.has(b) ? (e[b] = u, b in s && (s[b] = t)) : typeof u != "function" && Pi(e, b, u, o);
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
function Ii(e, t, n = [], r = [], i = [], a, o = !1, s = !1) {
	lt(i, n, r, (n) => {
		var r = void 0, i = {}, c = e.nodeName === ji, l = !1;
		if (kn(() => {
			var u = t(...n.map(J)), d = Fi(e, r, u, a, o, s);
			if (l && c) {
				var f = e;
				"defaultValue" in u && vi(f, u.defaultValue), "value" in u && bi(f, u.value);
			}
			for (let e of Object.getOwnPropertySymbols(i)) u[e] || B(i[e]);
			for (let t of Object.getOwnPropertySymbols(u)) {
				var p = u[t];
				t.description !== "@attach" || r && p === r[t] || (i[t] && B(i[t]), i[t] = z(() => ai(e, () => p))), d[t] = p;
			}
			r = d;
		}), c) {
			var u = e;
			wn(() => {
				var e = r;
				"defaultValue" in e && vi(u, e.defaultValue), bi(u, e.value, !0), xi(u);
			});
		}
		l = !0;
	});
}
function Li(e) {
	return e[he] ??= {
		[Ei]: e.nodeName.includes("-"),
		[Di]: e.namespaceURI === n
	};
}
var Ri = /* @__PURE__ */ new Map();
function zi(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Ri.get(t);
	if (n) return n;
	Ri.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = u(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = p(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Bi = !1;
function Vi(e) {
	var t = Bi;
	try {
		return Bi = !1, [e(), Bi];
	} finally {
		Bi = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
var Hi = {
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
function Z(e, t, n) {
	return new Proxy({
		props: e,
		exclude: t
	}, Hi);
}
var Ui = {
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
function Q(...e) {
	return new Proxy({ props: e }, Ui);
}
function Wi(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ pt(r), J(u)) : (c && (c = !1, s = o ? sr(r) : r), s);
	let f;
	if (a) {
		var p = de in e || pe in e;
		f = l(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Vi(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ve(t), f(m)));
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
	var v = !1, y = (n & 1 ? pt : gt)(() => (v = !1, g()));
	a && J(y);
	var b = W;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? J(y) : i && a ? $t(e) : e;
			return Gt(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Vn && v || b.f & 16384 ? y.v : J(y);
	});
}
// kit/plugin.svelte.js
function Gi(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ Ut($t({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return J(this.#e);
		}
		set props(e) {
			Gt(this.#e, e, !0);
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
			if (!document.querySelector(`link[href="${n}"]`)) {
				let e = document.createElement("link");
				e.rel = "stylesheet", e.href = n, document.head.append(e);
			}
			this.style.display = "contents", this.app ||= zr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && Ur(this.app), this.app = null;
		}
	});
}
function Ki(e, t) {
	let n = $t({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Xt(r);
		}
	}), r = /* @__PURE__ */ Ut(0);
	xn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Xt(r);
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
	return xn(() => {
		let n = e(), c = t();
		J(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function qi(e) {
	let t = $t({});
	return xn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/format.js
function Ji(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function Yi(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// node_modules/@lucide/svelte/dist/utils/defaultAttributes.js
var Xi = {
	xmlns: "http://www.w3.org/2000/svg",
	width: 24,
	height: 24,
	viewBox: "0 0 24 24",
	fill: "none",
	stroke: "currentColor",
	"stroke-width": 2,
	"stroke-linecap": "round",
	"stroke-linejoin": "round"
}, Zi = (...e) => e.filter((e, t, n) => !!e && e.trim() !== "" && n.indexOf(e) === t).join(" ").trim();
// node_modules/@lucide/svelte/dist/utils/buildLucideIconNode.js
function Qi(e) {
	return e != null;
}
function $i(e, t = {}) {
	let n = t.attributeNames ?? {}, r = (e) => n[e] ?? e, i = e.size ?? e.width ?? Xi.width, a = e.size ?? e.height ?? Xi.height, o = e.aliases?.filter((e) => typeof e == "string" && e.trim() !== "").map((e) => `lucide-${e}`) ?? [], s = [...e.name ? [`lucide-${e.name}`] : [], ...o], c = t.className?.split(" ").filter(Boolean) ?? [], l = t.includeDefaultClasses === !1 ? Zi(...c) : Zi("lucide", ...s, ...c), u = t.absoluteStrokeWidth ? Number(t.strokeWidth ?? Xi["stroke-width"]) * Number(e.size ?? e.width ?? Xi.width) / Number(t.size ?? t.width ?? Xi.width) : t.strokeWidth ?? Xi["stroke-width"];
	return [
		"svg",
		{
			...Object.entries(Xi).reduce((e, [t, n]) => (e[r(t)] = n, e), {}),
			..."color" in t && t.color && { [r("stroke")]: t.color },
			..."size" in t && Qi(t.size) && {
				[r("width")]: t.size,
				[r("height")]: t.size
			},
			..."width" in t && Qi(t.width) && { [r("width")]: t.width },
			..."height" in t && Qi(t.height) && { [r("height")]: t.height },
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
var ea = (e) => {
	for (let t in e) if (t.startsWith("aria-") || t === "role" || t === "title") return !0;
	return !1;
}, ta = Symbol("lucide-context"), na = () => Ye(ta), ra = /* @__PURE__ */ new Set([
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
]), ia = /* @__PURE__ */ Sr("<svg><!><!></svg>");
function $(e, t) {
	Xe(t, !0);
	let n = na() ?? {}, r = Wi(t, "color", 19, () => n.color ?? "currentColor"), i = Wi(t, "size", 19, () => n.size ?? 24), a = Wi(t, "width", 19, i), o = Wi(t, "height", 19, i), s = Wi(t, "strokeWidth", 19, () => n.strokeWidth ?? 2), c = Wi(t, "absoluteStrokeWidth", 19, () => n.absoluteStrokeWidth ?? !1), l = Wi(t, "nonScalingStroke", 19, () => n.nonScalingStroke ?? !1), u = Wi(t, "iconNode", 19, () => []), d = Wi(t, "icon", 19, () => ({
		node: u(),
		aliases: [],
		size: 24
	})), f = /* @__PURE__ */ Z(t, ra), p = /* @__PURE__ */ M(() => !!t.children || ea(f)), m = /* @__PURE__ */ M(() => $i(d(), {
		color: r(),
		width: a(),
		height: o(),
		strokeWidth: s(),
		absoluteStrokeWidth: c(),
		nonScalingStroke: l(),
		className: Zi("lucide-icon", n.class),
		hasA11yProp: J(p),
		attributes: f
	})), h = /* @__PURE__ */ M(() => b(J(m), 3)), _ = /* @__PURE__ */ M(() => J(h)[1]), v = /* @__PURE__ */ M(() => y(J(h)[2], () => [], !0)), x = /* @__PURE__ */ M(() => ({
		...J(_),
		class: [...J(_).class.split(" "), t.class]
	}));
	var S = ia();
	Ii(S, () => ({ ...J(x) }));
	var C = I(S);
	Zr(C, 17, () => J(v), qr, (e, t) => {
		var n = /* @__PURE__ */ M(() => b(J(t), 2));
		let r = () => J(n)[0], i = () => J(n)[1];
		var a = Cr();
		ii(ln(a), r, !0, (e, t) => {
			Ii(e, () => ({ ...i() }));
		}), X(e, a);
	}), Gr(R(C), () => t.children ?? g), O(S), X(e, S), Ze();
}
// node_modules/@lucide/svelte/dist/icons/git-commit-horizontal.svelte
var aa = {
	name: "git-commit-horizontal",
	size: 24,
	node: [
		["circle", {
			cx: "12",
			cy: "12",
			r: "3"
		}],
		["line", {
			x1: "3",
			x2: "9",
			y1: "12",
			y2: "12"
		}],
		["line", {
			x1: "15",
			x2: "21",
			y1: "12",
			y2: "12"
		}]
	],
	aliases: ["git-commit"]
};
aa.node;
var oa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function sa(e, t) {
	let n = /* @__PURE__ */ Z(t, oa);
	$(e, Q(() => n, { get icon() {
		return aa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-pull-request.svelte
var ca = {
	name: "git-pull-request",
	size: 24,
	node: [
		["circle", {
			cx: "18",
			cy: "18",
			r: "3"
		}],
		["circle", {
			cx: "6",
			cy: "6",
			r: "3"
		}],
		["path", { d: "M13 6h3a2 2 0 0 1 2 2v7" }],
		["line", {
			x1: "6",
			x2: "6",
			y1: "9",
			y2: "21"
		}]
	]
};
ca.node;
var la = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ua(e, t) {
	let n = /* @__PURE__ */ Z(t, la);
	$(e, Q(() => n, { get icon() {
		return ca;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-merge.svelte
var da = {
	name: "git-merge",
	size: 24,
	node: [
		["circle", {
			cx: "18",
			cy: "18",
			r: "3"
		}],
		["circle", {
			cx: "6",
			cy: "6",
			r: "3"
		}],
		["path", { d: "M6 21V9a9 9 0 0 0 9 9" }]
	]
};
da.node;
var fa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function pa(e, t) {
	let n = /* @__PURE__ */ Z(t, fa);
	$(e, Q(() => n, { get icon() {
		return da;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-branch.svelte
var ma = {
	name: "git-branch",
	size: 24,
	node: [
		["path", { d: "M15 6a9 9 0 0 0-9 9V3" }],
		["circle", {
			cx: "18",
			cy: "6",
			r: "3"
		}],
		["circle", {
			cx: "6",
			cy: "18",
			r: "3"
		}]
	]
};
ma.node;
var ha = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ga(e, t) {
	let n = /* @__PURE__ */ Z(t, ha);
	$(e, Q(() => n, { get icon() {
		return ma;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square.svelte
var _a = {
	name: "message-square",
	size: 24,
	node: [["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }]]
};
_a.node;
var va = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ya(e, t) {
	let n = /* @__PURE__ */ Z(t, va);
	$(e, Q(() => n, { get icon() {
		return _a;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square-code.svelte
var ba = {
	name: "message-square-code",
	size: 24,
	node: [
		["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }],
		["path", { d: "m10 8-3 3 3 3" }],
		["path", { d: "m14 14 3-3-3-3" }]
	]
};
ba.node;
var xa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Sa(e, t) {
	let n = /* @__PURE__ */ Z(t, xa);
	$(e, Q(() => n, { get icon() {
		return ba;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/eye.svelte
var Ca = {
	name: "eye",
	size: 24,
	node: [["path", { d: "M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" }], ["circle", {
		cx: "12",
		cy: "12",
		r: "3"
	}]]
};
Ca.node;
var wa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ta(e, t) {
	let n = /* @__PURE__ */ Z(t, wa);
	$(e, Q(() => n, { get icon() {
		return Ca;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/circle-check.svelte
var Ea = {
	name: "circle-check",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "12",
		r: "10"
	}], ["path", { d: "m16 9-5.5 5.5L8 12" }]],
	aliases: ["check-circle-2"]
};
Ea.node;
var Da = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Oa(e, t) {
	let n = /* @__PURE__ */ Z(t, Da);
	$(e, Q(() => n, { get icon() {
		return Ea;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/circle-x.svelte
var ka = {
	name: "circle-x",
	size: 24,
	node: [
		["circle", {
			cx: "12",
			cy: "12",
			r: "10"
		}],
		["path", { d: "m15 9-6 6" }],
		["path", { d: "m9 9 6 6" }]
	],
	aliases: ["x-circle"]
};
ka.node;
var Aa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ja(e, t) {
	let n = /* @__PURE__ */ Z(t, Aa);
	$(e, Q(() => n, { get icon() {
		return ka;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/circle-dot.svelte
var Ma = {
	name: "circle-dot",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "12",
		r: "1"
	}], ["circle", {
		cx: "12",
		cy: "12",
		r: "10"
	}]]
};
Ma.node;
var Na = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Pa(e, t) {
	let n = /* @__PURE__ */ Z(t, Na);
	$(e, Q(() => n, { get icon() {
		return Ma;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/tag.svelte
var Fa = {
	name: "tag",
	size: 24,
	node: [["path", { d: "M12.586 2.586A2 2 0 0 0 11.172 2H4a2 2 0 0 0-2 2v7.172a2 2 0 0 0 .586 1.414l8.704 8.704a2.426 2.426 0 0 0 3.42 0l6.58-6.58a2.426 2.426 0 0 0 0-3.42z" }], ["circle", {
		cx: "7.5",
		cy: "7.5",
		r: ".5",
		fill: "currentColor"
	}]]
};
Fa.node;
var Ia = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function La(e, t) {
	let n = /* @__PURE__ */ Z(t, Ia);
	$(e, Q(() => n, { get icon() {
		return Fa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/refresh-cw.svelte
var Ra = {
	name: "refresh-cw",
	size: 24,
	node: [
		["path", { d: "M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" }],
		["path", { d: "M21 3v5h-5" }],
		["path", { d: "M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" }],
		["path", { d: "M8 16H3v5" }]
	]
};
Ra.node;
var za = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ba(e, t) {
	let n = /* @__PURE__ */ Z(t, za);
	$(e, Q(() => n, { get icon() {
		return Ra;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/user-round.svelte
var Va = {
	name: "user-round",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "8",
		r: "5"
	}], ["path", { d: "M20 21a8 8 0 0 0-16 0" }]],
	aliases: ["user-2"]
};
Va.node;
var Ha = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ua(e, t) {
	let n = /* @__PURE__ */ Z(t, Ha);
	$(e, Q(() => n, { get icon() {
		return Va;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/activity.svelte
var Wa = {
	name: "activity",
	size: 24,
	node: [["path", { d: "M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" }]]
};
Wa.node;
var Ga = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ka(e, t) {
	let n = /* @__PURE__ */ Z(t, Ga);
	$(e, Q(() => n, { get icon() {
		return Wa;
	} }));
}
// plugins/activity/Activity.svelte
var qa = /* @__PURE__ */ Y("<button> </button>"), Ja = /* @__PURE__ */ Y("<option> </option>"), Ya = /* @__PURE__ */ Y("<div class=\"card err\"> </div>"), Xa = /* @__PURE__ */ Y("<div class=\"card empty faint svelte-dcm8ng\">Activity is off. Set <span class=\"mono\"></span> in jobshop.yaml.</div>"), Za = /* @__PURE__ */ Y("<span class=\"pill\">bot</span>"), Qa = /* @__PURE__ */ Y("<div class=\"body muted svelte-dcm8ng\"> </div>"), $a = /* @__PURE__ */ Y("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"svelte-dcm8ng\">open</a>"), eo = /* @__PURE__ */ Y("<a class=\"mono svelte-dcm8ng\"> </a>"), to = /* @__PURE__ */ Y("<li class=\"svelte-dcm8ng\"><span><!></span> <div class=\"main svelte-dcm8ng\"><div class=\"line svelte-dcm8ng\"><span class=\"actor svelte-dcm8ng\"> </span> <!> <span class=\"title svelte-dcm8ng\"> </span></div> <!> <div class=\"meta faint svelte-dcm8ng\"><span class=\"mono\"> </span> <span> </span> <!> <!></div></div></li>"), no = /* @__PURE__ */ Y("<h2 class=\"st\"> </h2> <ul class=\"card feed svelte-dcm8ng\" data-testid=\"activity\"></ul>", 1), ro = /* @__PURE__ */ Y("<div class=\"card empty faint svelte-dcm8ng\">Nothing in the last 7 days for these filters.</div>"), io = /* @__PURE__ */ Y("<!> <!>", 1), ao = /* @__PURE__ */ Y("<header class=\"vh svelte-dcm8ng\"><h1>Activity</h1> <div class=\"chips actions svelte-dcm8ng\" data-testid=\"activity-filters\"><!> <span class=\"sep\"></span> <button>all sources</button> <!> <span class=\"sep\"></span> <select aria-label=\"kind\"></select></div></header> <div class=\"pad\"><!></div>", 1);
function oo(e, t) {
	Xe(t, !0);
	let n = qi(() => t.api), r = (e) => t.api.go(e), i = [
		"",
		"push",
		"pr",
		"merge",
		"review",
		"review_comment",
		"comment",
		"issue",
		"ci",
		"branch",
		"release",
		"state",
		"assignee"
	], a = [
		"humans",
		"bots",
		"all"
	], o = /* @__PURE__ */ M(() => n.who || "humans"), s = Ki(() => t.api, () => {
		let e = new URLSearchParams({
			since: "168h",
			limit: "300"
		});
		return n.source && e.set("source", n.source), n.repo && e.set("repo", n.repo), n.kind && e.set("kind", n.kind), J(o) !== "all" && e.set("bots", J(o)), "api/activity?" + e.toString();
	}), c = (e) => {
		let t = e.indexOf(":");
		return {
			source: e.slice(0, t),
			repo: e.slice(t + 1)
		};
	};
	function l(e) {
		switch (e.kind) {
			case "push": return sa;
			case "pr": return ua;
			case "merge": return pa;
			case "branch": return ga;
			case "comment": return ya;
			case "review_comment": return Sa;
			case "review": return Ta;
			case "ci": return / success /.test(e.title) ? Oa : ja;
			case "issue": return Pa;
			case "release": return La;
			case "state": return Ba;
			case "assignee": return Ua;
			default: return Ka;
		}
	}
	let u = (e) => e.kind === "ci" ? / success /.test(e.title) ? "ok" : / (failure|timed_out|startup_failure) /.test(e.title) ? "bad" : "" : "";
	function d(e) {
		if (!e) return "";
		let t = e.replace(/\r/g, "").trim();
		return t.length > 280 ? t.slice(0, 280) + "…" : t;
	}
	function f(e) {
		let t = new Date(e), n = /* @__PURE__ */ new Date(), r = new Date(n);
		return r.setDate(n.getDate() - 1), t.toDateString() === n.toDateString() ? "Today" : t.toDateString() === r.toDateString() ? "Yesterday" : t.toLocaleDateString(void 0, {
			weekday: "short",
			month: "short",
			day: "numeric"
		});
	}
	let p = /* @__PURE__ */ M(() => {
		let e = [];
		for (let t of s.data?.items ?? []) {
			let n = f(t.at);
			e.length && e[e.length - 1].day === n || e.push({
				day: n,
				items: []
			}), e[e.length - 1].items.push(t);
		}
		return e;
	});
	var m = ao(), h = ln(m), g = R(I(h), 2), _ = I(g);
	Zr(_, 16, () => a, (e) => e, (e, t) => {
		var n = qa();
		let i;
		var a = L(n, !0);
		Dn(() => {
			i = mi(n, 1, "", null, i, { on: J(o) === t }), Rr(a, t);
		}), fr("click", n, () => r({ who: t === "humans" ? "" : t })), X(e, n);
	});
	var v = R(_, 4);
	let y;
	var b = R(v, 2);
	Zr(b, 16, () => s.data?.sources ?? [], (e) => e, (e, t) => {
		let i = /* @__PURE__ */ M(() => c(t));
		var a = qa();
		let o;
		var s = L(a, !0);
		Dn(() => {
			o = mi(a, 1, "", null, o, { on: n.repo === J(i).repo }), Rr(s, J(i).repo);
		}), fr("click", a, () => r({
			source: J(i).source,
			repo: J(i).repo
		})), X(e, a);
	});
	var x = R(b, 4);
	Zr(x, 20, () => i, (e) => e, (e, t) => {
		var n = Ja(), r = L(n, !0), i = {};
		Dn(() => {
			Rr(r, t || "all kinds"), i !== (i = t) && (n.value = (n.__value = i) ?? "");
		}), X(e, n);
	}), O(x);
	var S;
	xi(x), O(g), O(h);
	var C = R(h, 2), w = I(C), ee = (e) => {
		var t = Ya(), n = L(t, !0);
		Dn(() => Rr(n, s.error)), X(e, t);
	}, te = (e) => {
		var t = io(), n = ln(t), i = (e) => {
			var t = Xa(), n = R(I(t));
			n.textContent = "activity: {}", Oe(), O(t), X(e, t);
		};
		Kr(n, (e) => {
			s.data.sources.length || e(i);
		}), Zr(R(n, 2), 17, () => J(p), (e) => e.day, (e, t) => {
			var n = no(), i = ln(n), a = L(i, !0), o = R(i, 2);
			Zr(o, 21, () => J(t).items, (e) => e.id, (e, t) => {
				let n = /* @__PURE__ */ M(() => l(J(t)));
				var i = to(), a = I(i);
				ri(I(a), () => J(n), (e, t) => {
					t(e, { size: 14 });
				}), O(a);
				var o = R(a, 2), s = I(o), c = I(s), f = L(c, !0), p = R(c, 2), m = (e) => {
					X(e, Za());
				};
				Kr(p, (e) => {
					J(t).bot && e(m);
				});
				var h = L(R(p, 2), !0);
				O(s);
				var g = R(s, 2), _ = (e) => {
					var n = Qa(), r = L(n, !0);
					Dn((e) => Rr(r, e), [() => d(J(t).body)]), X(e, n);
				};
				Kr(g, (e) => {
					J(t).body && e(_);
				});
				var v = R(g, 2), y = I(v), b = L(y, !0), x = R(y, 2), S = L(x, !0), C = R(x, 2), w = (e) => {
					var n = $a();
					Dn(() => Pi(n, "href", J(t).url)), X(e, n);
				}, ee = /* @__PURE__ */ M(() => J(t).url && Yi(J(t).url));
				Kr(C, (e) => {
					J(ee) && e(w);
				}), Zr(R(C, 2), 16, () => J(t).tickets ?? [], (e) => e, (e, t) => {
					var n = eo(), i = L(n, !0);
					Dn(() => {
						Pi(n, "href", "?ticket=" + t), Rr(i, t);
					}), fr("click", n, (e) => {
						e.preventDefault(), r({ ticket: t });
					}), X(e, n);
				}), O(v), O(o), O(i), Dn((e, n) => {
					mi(a, 1, `icon ${e ?? ""}`, "svelte-dcm8ng"), Pi(a, "title", J(t).kind), Rr(f, J(t).actor || "unknown"), Rr(h, J(t).title), Rr(b, J(t).repo), Pi(x, "title", J(t).at), Rr(S, n);
				}, [() => u(J(t)), () => Ji(J(t).at)]), X(e, i);
			}), O(o), Dn(() => Rr(a, J(t).day)), X(e, n);
		}, (e) => {
			var t = Cr(), n = ln(t), r = (e) => {
				X(e, ro());
			};
			Kr(n, (e) => {
				s.data.sources.length && e(r);
			}), X(e, t);
		}), X(e, t);
	};
	Kr(w, (e) => {
		s.error ? e(ee) : s.data && e(te, 1);
	}), O(C), Dn(() => {
		y = mi(v, 1, "", null, y, { on: !n.source && !n.repo }), S !== (S = n.kind) && (x.value = (x.__value = S) ?? "", bi(x, S));
	}), fr("click", v, () => r({
		source: "",
		repo: ""
	})), fr("change", x, (e) => r({ kind: e.currentTarget.value })), X(e, m), Ze();
}
// plugins/activity/index.js
pr(["click", "change"]), Gi("desk-activity-view", oo);
