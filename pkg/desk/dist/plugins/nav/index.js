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
function E(e) {
	T = e;
}
var D;
function O(t) {
	if (t === null) throw Ce(), e;
	return D = t;
}
function Ee() {
	return O(/* @__PURE__ */ rn(D));
}
function De(t) {
	if (T) {
		if (/* @__PURE__ */ rn(D) !== null) throw Ce(), e;
		D = t;
	}
}
function Oe(e = 1) {
	if (T) {
		for (var t = e, n = D; t--;) n = /* @__PURE__ */ rn(n);
		D = n;
	}
}
function ke(e = !0) {
	for (var t = 0, n = D;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ rn(n);
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
		for (var r of n) xn(r);
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
	if (et.length === 0 && !Tt) {
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
	U(null), Hn(null);
	try {
		return e();
	} finally {
		U(t), Hn(n);
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
				mn(e, s);
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
		Promise.all(n.map((e) => /* @__PURE__ */ ht(e))).then(u).catch((e) => mn(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), dt());
	}) : f();
}
function ut() {
	var e = W, t = V, n = k, r = N;
	return function(i = !0) {
		Hn(e), U(t), Je(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function dt(e = !0) {
	Hn(null), U(null), Je(null), e && N?.deactivate();
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
	var a = void 0, o = Vt(t), s = !V, c = /* @__PURE__ */ new Set();
	return wn(() => {
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
			l?.(), c.delete(n), t !== mt && (r.activate(), t ? (o.f |= ue, Gt(o, t)) : (o.f & 8388608 && (o.f ^= ue), Gt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), yn(() => {
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
	return Wn(t), t;
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
	if (!Bn && i !== null && e.v !== t && i.f & 24576) return Se(), e.v;
	Hn(i);
	try {
		_t(e), n = $n(e);
	} finally {
		Hn(r);
	}
	return n;
}
function yt(e) {
	var t = vt(e);
	!e.equals(t) && (e.wv = Xn(), (!N?.is_fork || e.deps === null) && (N === null ? e.v = t : (N.capture(e, t, !0), Ct?.capture(e, t, !0)), e.deps === null)) ? j(e, S) : Bn || (P === null ? rt(e) : (vn() || N?.is_fork) && P.set(e, t));
}
function bt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && ct(() => {
		t.ac.abort(be), t.ac = null;
	}), t.fn !== null && (t.teardown = g), nr(t, 0), An(t));
}
function xt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && rr(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var St = null, N = null, Ct = null, P = null, wt = null, Tt = !1, Et = !1, Dt = null, Ot = null, kt = 0, At = 1, jt = class e {
	id = At++;
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
		for (var t = Dt = [], n = [], r = Ot = []; this.#a.length > 0;) {
			kt++ > 1e3 && (this.#b(), Mt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Lt(e), this.#d() || this.discard(), t;
			}
		}
		if (N = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (Dt = null, Ot = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) It(e, t);
			r.length > 0 && N.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), Ct = this, Pt(n), Pt(t), Ct = null, this.#i?.resolve();
			var o = N;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (zt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= S;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= S : i & 4 ? t.push(r) : Zn(r) && (i & 16 && this.add_dirty_reaction(r, w), rr(r));
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
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), P?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		N = this;
	}
	deactivate() {
		N = null, P = null;
	}
	flush() {
		try {
			Et = !0, N = this, this.#m();
		} finally {
			kt = 0, wt = null, Dt = null, Ot = null, Et = !1, N = null, P = null, zt.clear();
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
			!Et && A(() => {
				t.started || t.flush();
			});
		}
		return N;
	}
	apply() {
		P = null;
	}
	schedule(e) {
		wt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? St = e : t.prev = e, this.linked = !1;
		}
	}
};
function Mt() {
	try {
		Be();
	} catch (e) {
		mn(e, wt);
	}
}
var Nt = null;
function Pt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Zn(r) && (Nt = /* @__PURE__ */ new Set(), rr(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && Nn(r), Nt?.size > 0)) {
				zt.clear();
				for (let e of Nt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Nt.has(n) && (Nt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || rr(n);
					}
				}
				Nt.clear();
			}
		}
		Nt = null;
	}
}
function Ft(e) {
	N.schedule(e);
}
function It(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), j(e, S);
		for (var n = e.first; n !== null;) It(n, t), n = n.next;
	}
}
function Lt(e) {
	j(e, S);
	for (var t = e.first; t !== null;) Lt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Rt = /* @__PURE__ */ new Set(), zt = /* @__PURE__ */ new Map(), Bt = !1;
function Vt(e, t) {
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
function F(e, t) {
	let n = Vt(e, t);
	return Wn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Ht(e, t = !1, n = !0) {
	let r = Vt(e);
	return t || (r.equals = Ne), r;
}
function I(e, t, n = !1) {
	return V !== null && (!H || V.f & 131072) && $e() && V.f & 4325394 && (Un === null || !Un.has(e)) && We(), Gt(e, n ? Yt(t) : t, Ot);
}
var Ut = null, Wt = 0;
function Gt(e, t, n = null) {
	if (!e.equals(t)) {
		Bn ? zt.set(e, t) : zt.has(e) || zt.set(e, e.v);
		var r = jt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && vt(t), P === null && rt(t);
		}
		e.wv = Xn(), Ut = null, Wt = 0, Jt(e, C, n), Ut = null, $e() && W !== null && W.f & 1024 && !(W.f & 96) && (q === null ? Gn([e]) : q.push(e)), !r.is_fork && Rt.size > 0 && !Bt && Kt();
	}
	return t;
}
function Kt() {
	Bt = !1;
	for (let e of Rt) {
		e.f & 1024 && j(e, w);
		let t;
		try {
			t = Zn(e);
		} catch {
			t = !0;
		}
		t && rr(e);
	}
	Rt.clear();
}
function qt(e) {
	I(e, e.v + 1);
}
function Jt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = $e(), a = r.length;
		if (Wt += a, Wt > 1e5 && Ut === null && (Ut = /* @__PURE__ */ new Set()), Ut !== null) {
			if (Ut.has(e)) return;
			Ut.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & C) === 0;
				if (l && j(s, t), c & 131072) Rt.add(s);
				else if (c & 2) {
					var u = s;
					P?.delete(u), Jt(u, w, n);
				} else if (l) {
					var d = s;
					c & 16 && Nt !== null && Nt.add(d), n === null ? Ft(d) : n.push(d);
				}
			}
		}
	}
}
function Yt(e) {
	if (typeof e != "object" || !e || de in e || fe in e) return e;
	let n = p(e);
	if (n !== d && n !== f) return e;
	var r = /* @__PURE__ */ new Map(), a = i(e), o = /* @__PURE__ */ F(0), s = null, c = Jn, u = (e) => {
		if (Jn === c) return e();
		var t = V, n = Jn;
		U(null), Yn(c);
		var r = e();
		return U(t), Yn(n), r;
	};
	return a && r.set("length", /* @__PURE__ */ F(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || He();
			var i = r.get(t);
			return i === void 0 ? u(() => {
				var e = /* @__PURE__ */ F(n.value, s);
				return r.set(t, e), e;
			}) : I(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = u(() => /* @__PURE__ */ F(t, s));
					r.set(n, e), qt(o);
				}
			} else I(i, t), qt(o);
			return !0;
		},
		get(n, i, a) {
			if (i === de) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || l(n, i)?.writable) && (o = u(() => /* @__PURE__ */ F(Yt(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
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
			return (i !== void 0 || (W !== null || V !== null) && (!a || l(e, n)?.writable)) && (i === void 0 && (i = u(() => /* @__PURE__ */ F(a ? Yt(e[n]) : t, s)), r.set(n, i)), J(i) === t) ? !1 : a;
		},
		set(e, n, i, c) {
			var d = r.get(n), f = n in e;
			if (a && n === "length") for (var p = i; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = u(() => /* @__PURE__ */ F(t, s)), r.set(p + "", m)) : I(m, t);
			}
			if (d === void 0) (!f || l(e, n)?.writable) && (d = u(() => /* @__PURE__ */ F(void 0, s)), I(d, Yt(i)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = u(() => Yt(i));
				I(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, i), !f) {
				if (a && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && I(_, v + 1);
				}
				qt(o);
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
function Xt(e) {
	try {
		if (typeof e == "object" && e && de in e) return e[de];
	} catch {}
	return e;
}
function Zt(e, t) {
	return Object.is(Xt(e), Xt(t));
}
var Qt, $t, en, tn;
function nn() {
	if (Qt === void 0) {
		Qt = window, $t = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		en = l(t, "firstChild").get, tn = l(t, "nextSibling").get, m(e) && (e[ge] = void 0, e[he] = null, e[_e] = void 0, e.__e = void 0), m(n) && (n[ve] = void 0);
	}
}
function L(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function R(e) {
	return en.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function rn(e) {
	return tn.call(e);
}
function an(e, t) {
	if (!T) return /* @__PURE__ */ R(e);
	var n = /* @__PURE__ */ R(D);
	if (n === null) n = D.appendChild(L());
	else if (t && n.nodeType !== 3) {
		var r = L();
		return n?.before(r), O(r), r;
	}
	return t && fn(n), O(n), n;
}
function on(e, t = !1) {
	if (!T) {
		var n = /* @__PURE__ */ R(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ rn(n) : n;
	}
	if (t) {
		if (D?.nodeType !== 3) {
			var r = L();
			return D?.before(r), O(r), r;
		}
		fn(D);
	}
	return D;
}
function sn(e, t = !1) {
	if (!T) return /* @__PURE__ */ R(e);
	var n = an(e, t);
	return De(e), n;
}
function cn(e, t = 1, n = !1) {
	let r = T ? D : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ rn(r);
	if (!T) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = L();
			return r === null ? i?.after(a) : r.before(a), O(a), a;
		}
		fn(r);
	}
	return O(r), r;
}
function ln(e) {
	e.textContent = "";
}
function un() {
	return !1;
}
function dn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function fn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function pn(e) {
	var t = W;
	if (t === null) return V.f |= ue, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	mn(e, t);
}
function mn(t, n) {
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
function hn(e) {
	W === null && (V === null && ze(e), Re()), Bn && Le(e);
}
function gn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function _n(e, t) {
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
	if (e & 4) Dt === null ? jt.ensure().schedule(r) : Dt.push(r);
	else if (t !== null) {
		try {
			rr(r);
		} catch (e) {
			throw B(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ie));
	}
	if (i !== null && (i.parent = n, n !== null && gn(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function vn() {
	return V !== null && !H;
}
function yn(e) {
	let t = _n(8, null);
	return j(t, S), t.teardown = e, t;
}
function bn(e) {
	hn("$effect");
	var t = W.f;
	if (!V && t & 32 && k !== null && !k.i) {
		var n = k;
		(n.e ??= []).push(e);
	} else return xn(e);
}
function xn(e) {
	return _n(4 | oe, e);
}
function Sn(e) {
	jt.ensure();
	let t = _n(64 | ae, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? Pn(t, () => {
			B(t), n(void 0);
		}) : (B(t), n(void 0));
	});
}
function Cn(e) {
	return _n(4, e);
}
function wn(e) {
	return _n(le | ae, e);
}
function Tn(e, t = 0) {
	return _n(8 | t, e);
}
function En(e, t = [], n = [], r = []) {
	lt(r, t, n, (t) => {
		_n(8, () => {
			e(...t.map(J));
		});
	});
}
function Dn(e, t = 0) {
	return _n(16 | t, e);
}
function On(e, t = 0) {
	return _n(x | t, e);
}
function z(e) {
	return _n(32 | ae, e);
}
function kn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Bn, r = V;
		Vn(!0), U(null);
		try {
			t.call(null);
		} catch (t) {
			mn(t, e.parent);
		} finally {
			Vn(n), U(r);
		}
	}
}
function An(e, t = !1) {
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
function jn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || B(t), t = n;
	}
}
function B(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Mn(e.nodes.start, e.nodes.end), n = !0), e.f |= re, An(e, t && !n), nr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	kn(e), e.f ^= re, e.f |= te;
	var i = e.parent;
	i !== null && i.first !== null && Nn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Mn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ rn(e);
		e.remove(), e = n;
	}
}
function Nn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function Pn(e, t, n = !0) {
	var r = [];
	e.f |= 256, Fn(e, r, !0);
	var i = () => {
		n && B(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function Fn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ee;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				Fn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function In(e) {
	e.f &= -257, Ln(e, !0);
}
function Ln(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ee, e.f & 1024 || (j(e, C), jt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Ln(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Rn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ rn(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var zn = null, Bn = !1;
function Vn(e) {
	Bn = e;
}
var V = null, H = !1;
function U(e) {
	V = e;
}
var W = null;
function Hn(e) {
	W = e;
}
var Un = null;
function Wn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (Un ??= /* @__PURE__ */ new Set()).add(e);
}
var G = null, K = 0, q = null;
function Gn(e) {
	q = e;
}
var Kn = 1, qn = 0, Jn = qn;
function Yn(e) {
	Jn = e;
}
function Xn() {
	return ++Kn;
}
function Zn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Zn(a) && yt(a), a.wv > e.wv) return !0;
		}
		t & 512 && P === null && j(e, S);
	}
	return !1;
}
function Qn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Un !== null && Un.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Qn(a, t, !1) : t === a && (n ? j(a, C) : a.f & 1024 && j(a, w), Ft(a));
	}
}
function $n(e) {
	var t = G, n = K, r = q, i = V, a = Un, o = k, s = H, c = Jn, l = e.f;
	G = null, K = 0, q = null, V = l & 96 ? null : e, Un = null, Je(e.ctx), H = !1, Jn = ++qn, e.ac !== null && (ct(() => {
		e.ac.abort(be);
	}), e.ac = null);
	try {
		e.f |= ce;
		var u = e.fn, d = u();
		e.f |= ne;
		var f = er(e);
		if ($e() && q !== null && !H && f !== null && !(e.f & 6146)) for (var p = 0; p < q.length; p++) Qn(q[p], e);
		if (i !== null && i !== e) {
			if (qn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = qn;
			if (t !== null) for (let e of t) e.rv = qn;
			q !== null && (r === null ? r = q : r.push(...q));
		}
		return e.f & 8388608 && (e.f ^= ue), d;
	} catch (t) {
		return er(e), pn(t);
	} finally {
		e.f ^= ce, G = t, K = n, q = r, V = i, Un = a, Je(o), H = s, Jn = c;
	}
}
function er(e) {
	var t = e.deps, n = N?.is_fork;
	if (G !== null) {
		var r;
		if (n || nr(e, K), t !== null && K > 0) for (t.length = K + G.length, r = 0; r < G.length; r++) t[K + r] = G[r];
		else e.deps = t = G;
		if (vn() && e.f & 512) for (r = K; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && K < t.length && (nr(e, K), t.length = K);
	return t;
}
function tr(e, n) {
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
		}), bt(c), nr(c, 0);
	}
}
function nr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) tr(e, n[r]);
}
function rr(e) {
	var t = e.f;
	if (!(t & 16384)) {
		j(e, S);
		var n = W;
		W = e;
		try {
			t & 16777232 ? jn(e) : An(e), kn(e);
			var r = $n(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Kn;
		} finally {
			W = n;
		}
	}
}
function J(e) {
	var t = !!(e.f & 2);
	if (zn?.add(e), V !== null && !H && !(W !== null && W.f & 16384) && (Un === null || !Un.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < qn && (e.rv = qn, G === null && n !== null && n[K] === e ? K++ : G === null ? G = [e] : G.push(e));
		else {
			V.deps ??= [], o.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : o.call(r, V) || r.push(V);
		}
	}
	if (Bn && zt.has(e)) return zt.get(e);
	if (t) {
		var i = e;
		if (Bn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || ar(i)) && (a = vt(i)), zt.set(i, a), a;
		}
		var s = !(i.f & 512) && !H && V !== null && !!(V.f & 512), c = (i.f & ne) === 0;
		Zn(i) && (s && (i.f |= 512), yt(i)), s && !c && (xt(i), ir(i));
	}
	if (P?.has(e)) return P.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function ir(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : o.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (xt(n), ir(n));
	}
}
function ar(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (zt.has(t) || t.f & 2 && ar(t)) return !0;
	return !1;
}
function or(e) {
	var t = H;
	try {
		return H = !0, e();
	} finally {
		H = t;
	}
}
function sr(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (de in e) cr(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && de in n && cr(n);
		}
	}
}
function cr(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			cr(e[n], t);
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
var lr = Symbol("events"), ur = /* @__PURE__ */ new Set(), dr = /* @__PURE__ */ new Set();
function fr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || vr.call(t, e), !e.cancelBubble) return ct(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, A(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function pr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = fr(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && yn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function mr(e, t, n) {
	(t[lr] ??= {})[e] = n;
}
function hr(e) {
	for (var t = 0; t < e.length; t++) ur.add(e[t]);
	for (var n of dr) n(e);
}
var gr = null, _r = !1;
function vr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	gr = e, _r || (_r = !0, setTimeout(() => {
		_r = !1, gr = null;
	}));
	var o = 0, s = gr === e && e[lr];
	if (s) {
		var l = i.indexOf(s);
		if (l !== -1 && (t === document || t === window)) {
			e[lr] = t;
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
		U(null), Hn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[lr]?.[r];
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
			e[lr] = t, delete e.currentTarget, U(d), Hn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var yr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function br(e) {
	return yr?.createHTML(e) ?? e;
}
function xr(e) {
	var t = dn("template");
	return t.innerHTML = br(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Sr(e, t) {
	var n = W;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Cr(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (T) return Sr(D, null), D;
		i === void 0 && (i = xr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ R(i)));
		var t = r || $t ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ R(t), s = t.lastChild;
			Sr(o, s);
		} else Sr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function wr(e, t, n = "svg") {
	var r = !e.startsWith("<!>"), i = !!(t & 1), a = `<${n}>${r ? e : "<!>" + e}</${n}>`, o;
	return () => {
		if (T) return Sr(D, null), D;
		if (!o) {
			var e = /* @__PURE__ */ R(xr(a));
			if (i) for (o = document.createDocumentFragment(); /* @__PURE__ */ R(e);) o.appendChild(/* @__PURE__ */ R(e));
			else o = /* @__PURE__ */ R(e);
		}
		var t = o.cloneNode(!0);
		if (i) {
			var n = /* @__PURE__ */ R(t), r = t.lastChild;
			Sr(n, r);
		} else Sr(t, t);
		return t;
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Tr(e, t) {
	return /* @__PURE__ */ wr(e, t, "svg");
}
function Er() {
	if (T) return Sr(D, null), D;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = L();
	return e.append(t, n), Sr(t, n), e;
}
function Dr(e, t) {
	if (T) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = D), Ee();
	} else e !== null && e.before(t);
}
// node_modules/svelte/src/utils.js
function Or(e) {
	return e.endsWith("capture") && e !== "gotpointercapture" && e !== "lostpointercapture";
}
var kr = [
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
function Ar(e) {
	return kr.includes(e);
}
var jr = /* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split("."), Mr = {
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
function Nr(e) {
	return e = e.toLowerCase(), Mr[e] ?? e;
}
[...jr];
var Pr = ["touchstart", "touchmove"];
function Fr(e) {
	return Pr.includes(e);
}
var Ir = [
	"textarea",
	"script",
	"style",
	"title"
];
function Lr(e) {
	return Ir.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Rr(e) {
	let t = 0, n = Vt(0), r;
	return () => {
		vn() && (J(n), Tn(() => (t === 0 && (r = or(() => e(() => qt(n)))), t += 1, () => {
			A(() => {
				--t, t === 0 && (r?.(), r = void 0, qt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var zr = ie | ae;
function Br(e, t, n, r) {
	new Vr(e, t, n, r);
}
var Vr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = T ? D : null;
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
	#m = Rr(() => (this.#p = Vt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = Dn(() => {
			if (T) {
				let e = Ae(this.#t);
				Ee();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, zr), T && (this.#e = D);
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
			this.#v() || (t ? Te() : (t = !0, n && Ge(), this.#s !== null && Pn(this.#s, () => {
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
					mn(e, this.#i && this.#i.parent);
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
				var e = this.#c = document.createDocumentFragment(), t = L(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return z(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							mn(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(N)) : this.#u === 0 && (this.#e.before(e), this.#c = null, Pn(this.#o, () => {
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
				Rn(this.#a, e);
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
		Hn(this.#i), U(this.#i), Je(this.#i.ctx);
		try {
			return jt.ensure(), e();
		} finally {
			Hn(t), U(n), Je(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && Pn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, A(() => {
			this.#d = !1, this.#p && Gt(this.#p, this.#l);
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
		this.#a &&= (B(this.#a), null), this.#o &&= (B(this.#o), null), this.#s &&= (B(this.#s), null), T && (O(this.#t), Oe(), O(ke()));
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
					return mn(e, this.#i.parent), null;
				}
			}));
		};
		A(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					mn(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => mn(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Hr(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[ve] ??= e.nodeValue) && (e[ve] = n, e.nodeValue = `${n}`);
}
function Ur(e, t) {
	return Gr(e, t);
}
var Wr = /* @__PURE__ */ new Map();
function Gr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: c = !0, transformError: l }) {
	nn();
	var u = void 0, d = Sn(() => {
		var c = r ?? n.appendChild(L());
		Br(c, { pending: () => {} }, (n) => {
			Xe({});
			var r = k;
			if (o && (r.c = o), a && (i.$$events = a), T && Sr(n, null), u = t(n, i) || Qe(), T && (W.nodes.end = D, D === null || D.nodeType !== 8 || D.data !== "]")) throw Ce(), e;
			Ze();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Fr(r);
					for (let e of [n, document]) {
						var a = Wr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Wr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, vr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(s(ur)), dr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Wr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, vr), t.delete(e), t.size === 0 && Wr.delete(r)) : t.set(e, i);
			}
			dr.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return Kr.set(u, d), u;
}
var Kr = /* @__PURE__ */ new WeakMap();
function qr(e, t) {
	let n = Kr.get(e);
	return n ? (Kr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Jr = class {
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
			if (n) In(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (In(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
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
						Rn(r, t), t.append(L()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else B(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), Pn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (B(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = N, r = un();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = L();
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
		} else T && (this.anchor = D), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/snippet.js
function Yr(e, t, ...n) {
	var r = new Jr(e);
	Dn(() => {
		let e = t() ?? null;
		r.ensure(e, e && ((t) => e(t, ...n)));
	}, ie);
}
function Xr(e) {
	k === null && Pe("onMount"), bn(() => {
		let t = or(e);
		if (typeof t == "function") return t;
	});
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function Zr(e, t, n = !1) {
	var r;
	T && (r = D, Ee());
	var i = new Jr(e), a = n ? ie : 0;
	function o(e, t) {
		if (T) {
			var n = Ae(r);
			if (e !== parseInt(n.substring(1))) {
				var a = ke();
				O(a), i.anchor = a, E(!1), i.ensure(e, t), E(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	Dn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function Qr(e, t) {
	return t;
}
function $r(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, c = 0; c < i; c++) {
		let n = t[c];
		Pn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					ei(e, s(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			ln(d), d.append(u), e.items.clear();
		}
		ei(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function ei(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= se, Rn(a, document.createDocumentFragment())) : B(t[i], n);
	}
}
var ti;
function ni(e, t, n, r, a, o = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = T ? O(/* @__PURE__ */ R(u)) : u.appendChild(L());
	}
	T && Ee();
	var d = null, f = /* @__PURE__ */ gt(() => {
		var e = n();
		return i(e) ? e : e == null ? [] : s(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = J(f);
			_.fallback = d, ii(_, n, c, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= se, oi(d, null, c)) : In(d) : Pn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: Dn(() => {
			var e = J(f), i = e.length;
			let s = !1;
			T && Ae(c) === "[!" != (i === 0) && (c = ke(), O(c), E(!1), s = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = N, v = un(), y = 0; y < i; y += 1) {
				T && D.nodeType === 8 && D.data === "]" && (c = D, s = !0, E(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Gt(S.v, b), S.i && Gt(S.i, y), v && _.unskip_effect(S.e)) : (S = ai(l, m ? c : ti ??= L(), b, x, y, a, t, n), m || (S.e.f |= se), l.set(x, S)), u.add(x);
			}
			if (i === 0 && o && !d && (m ? d = z(() => o(c)) : (d = z(() => o(ti ??= L())), d.f |= se)), i > u.size && Ie("", "", ""), T && i > 0 && O(ke()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			s && E(!0), J(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, T && (c = D);
}
function ri(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function ii(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, c = e.items, l = ri(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (In(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= se, _ === l) oi(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), si(e, d, _), si(e, _, y), oi(_, y, n), d = _, p = [], m = [], l = ri(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) oi(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					si(e, S.prev, C.next), si(e, d, S), si(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), oi(_, l, n), si(e, _.prev, _.next), si(e, _, d === null ? e.effect.first : d.next), si(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = ri(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = ri(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (ei(e, s(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = ri(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			$r(e, w, te);
		}
	}
	a && A(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function ai(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Vt(n) : /* @__PURE__ */ Ht(n, !1, !1) : null, l = o & 2 ? Vt(i) : null;
	return {
		v: c,
		i: l,
		e: z(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function oi(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ rn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function si(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-component.js
function ci(e, t, n) {
	var r;
	T && (r = D, Ee());
	var i = new Jr(e);
	Dn(() => {
		var e = t() ?? null;
		if (T && Ae(r) === "[" != (e !== null)) {
			var a = ke();
			O(a), i.anchor = a, E(!1), i.ensure(e, e && ((t) => n(t, e))), E(!0);
		} else i.ensure(e, e && ((t) => n(t, e)));
	}, ie);
}
// node_modules/svelte/src/internal/client/dom/blocks/svelte-element.js
function li(e, t, n, i, a, o) {
	let s = T;
	T && Ee();
	var c = null;
	T && D.nodeType === 1 && (c = D, Ee());
	var l = T ? D : e, u = new Jr(l, !1);
	Dn(() => {
		let e = t() || null;
		var o = a ? a() : n || e === "svg" ? r : void 0;
		if (e === null) u.ensure(null, null);
		else return u.ensure(e, (t) => {
			if (e) {
				var n = T;
				if (c = T ? c : dn(e, o), Sr(c, c), i) {
					var r = null;
					T && Lr(e) && c.append(r = document.createComment(""));
					var a = T ? /* @__PURE__ */ R(c) : c.appendChild(L());
					T && (a === null ? E(!1) : O(a)), i(c, a), r?.remove();
				}
				W.nodes.end = c, n || t.before(c);
			}
			T && O(t);
		}), () => {};
	}, ie), yn(() => {}), s && (E(!0), O(l));
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function ui(e, t, n) {
	Cn(() => {
		var r = or(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			Tn(() => {
				var e = n();
				sr(e), i && Me(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/client/dom/elements/attachments.js
function di(e, t) {
	var n = void 0, r;
	On(() => {
		n !== (n = t()) && (r &&= (B(r), null), n && (r = z(() => {
			Cn(() => n(e));
		})));
	});
}
// node_modules/clsx/dist/clsx.mjs
function fi(e) {
	var t, n, r = "";
	if (typeof e == "string" || typeof e == "number") r += e;
	else if (typeof e == "object") {
		if (Array.isArray(e)) {
			var i = e.length;
			for (t = 0; t < i; t++) e[t] && (n = fi(e[t])) && (r && (r += " "), r += n);
		} else for (n in e) e[n] && (r && (r += " "), r += n);
	}
	return r;
}
function pi() {
	for (var e, t, n = 0, r = "", i = arguments.length; n < i; n++) (e = arguments[n]) && (t = fi(e)) && (r && (r += " "), r += t);
	return r;
}
// node_modules/svelte/src/internal/shared/attributes.js
function mi(e) {
	return typeof e == "object" ? pi(e) : e ?? "";
}
var hi = [..." 	\n\r\f\xA0\v﻿"];
function gi(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || hi.includes(r[o - 1])) && (s === r.length || hi.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
function _i(e, t = !1) {
	var n = t ? " !important;" : ";", r = "";
	for (var i of Object.keys(e)) {
		var a = e[i];
		a != null && a !== "" && (r += " " + i + ": " + a + n);
	}
	return r;
}
function vi(e) {
	return e[0] !== "-" || e[1] !== "-" ? e.toLowerCase() : e;
}
function yi(e, t) {
	if (t) {
		var n = "", r, i;
		if (Array.isArray(t) ? (r = t[0], i = t[1]) : r = t, e) {
			e = String(e).replaceAll(/\/\*.*?\*\//g, "").trim();
			var a = !1, o = 0, s = !1, c = [];
			r && c.push(...Object.keys(r).map(vi)), i && c.push(...Object.keys(i).map(vi));
			var l = 0, u = -1;
			let t = e.length;
			for (var d = 0; d < t; d++) {
				var f = e[d];
				if (s ? f === "/" && e[d - 1] === "*" && (s = !1) : a ? a === f && (a = !1) : f === "/" && e[d + 1] === "*" ? s = !0 : f === "\"" || f === "'" ? a = f : f === "(" ? o++ : f === ")" && o--, !s && a === !1 && o === 0) {
					if (f === ":" && u === -1) u = d;
					else if (f === ";" || d === t - 1) {
						if (u !== -1) {
							var p = vi(e.substring(l, u).trim());
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
		return r && (n += _i(r)), i && (n += _i(i, !0)), n = n.trim(), n === "" ? null : n;
	}
	return e == null ? null : String(e);
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function bi(e, t, n, r, i, a) {
	var o = e[ge];
	if (T || o !== n || o === void 0) {
		var s = gi(n, r, a);
		T && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[ge] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/style.js
function xi(e, t = {}, n, r) {
	for (var i in n) {
		var a = n[i];
		t[i] !== a && (n[i] == null ? e.style.removeProperty(i) : e.style.setProperty(i, a, r));
	}
}
function Si(e, t, n, r) {
	var i = e[_e];
	if (T || i !== t) {
		var a = yi(t, r);
		T && a === e.getAttribute("style") || (a == null ? e.removeAttribute("style") : e.style.cssText = a), e[_e] = t;
	} else r && (Array.isArray(r) ? (xi(e, n?.[0], r[0]), xi(e, n?.[1], r[1], "important")) : xi(e, n, r));
	return r;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function Ci(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function wi(e, t) {
	var n = !("__defaultValue" in e);
	(n || e.__defaultValue !== t) && (e.__defaultValue = t, Ti(e, !n || "__value" in e));
}
function Ti(e, t) {
	var n = e.__defaultValue, r = e.multiple, a = r ? n ?? [] : null;
	if (!r || i(a)) {
		var o = e.selectedIndex, s = t && r ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = Oi(c);
			Ci(c, r ? a.includes(l) : Zt(l, n));
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
function Ei(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!i(t)) return we();
		for (var r of e.options) r.selected = t.includes(Oi(r));
	} else {
		for (r of e.options) if (Zt(Oi(r), t)) {
			r.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function Di(e) {
	var t = new MutationObserver((t) => {
		t.every(ki) || ("__defaultValue" in e && Ti(e, !1), "__value" in e && Ei(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), yn(() => {
		t.disconnect();
	});
}
function Oi(e) {
	return "__value" in e ? e.__value : e.value;
}
function ki(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Ai = Symbol("class"), ji = Symbol("style"), Mi = Symbol("is custom element"), Ni = Symbol("is html"), Pi = xe ? "link" : "LINK", Fi = xe ? "input" : "INPUT", Ii = xe ? "option" : "OPTION", Li = xe ? "select" : "SELECT", Ri = xe ? "progress" : "PROGRESS";
function zi(e) {
	if (T) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Y(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Y(e, "checked", null), e.checked = r;
				}
			}
		};
		e[ye] = n, A(n), st();
	}
}
function Y(e, t, n, r) {
	var i = Hi(e);
	if (!(T && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Pi))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[me] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Wi(e).has(t) ? e[t] = n : (!T || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Bi(e, n, r, i, a = !1, o = !1) {
	T && a && e.nodeName === Fi && ("defaultValue" in r || "defaultChecked" in r || zi(e));
	var s = Hi(e), c = s[Mi], l = !s[Ni];
	let u = T && c;
	u && E(!1);
	var d = n || {}, f = e.nodeName === Ii, p = e.nodeName === Li;
	try {
		for (var m in n) !(m in r) && m[0] + m[1] !== "$$" && (r[m] = null);
		r.class ? r.class = mi(r.class) : (i || r[Ai]) && (r.class = null), r[ji] && (r.style ??= null);
		var h = Wi(e);
		if (e.nodeName === Fi && "type" in r && ("value" in r || "__value" in r)) {
			var g = r.type;
			(g !== d.type || g === void 0 && e.hasAttribute("type")) && (d.type = g, Y(e, "type", g, o));
		}
		for (let a in r) {
			let u = r[a];
			if (f && a === "value" && u == null) e.value = e.__value = "", d[a] = u;
			else if (a === "class") bi(e, e.namespaceURI === "http://www.w3.org/1999/xhtml", u, i, n?.[Ai], r[Ai]), d[a] = u, d[Ai] = r[Ai];
			else if (a === "style") Si(e, u, n?.[ji], r[ji]), d[a] = u, d[ji] = r[ji];
			else {
				var _ = d[a];
				if (u !== _ || u === void 0 && e.hasAttribute(a)) {
					d[a] = u;
					var v = a[0] + a[1];
					if (v !== "$$") {
						if (v === "on") {
							let t = {}, n = "$$" + a, r = a.slice(2);
							var y = Ar(r);
							if (Or(r) && (r = r.slice(0, -7), t.capture = !0), !y && _) {
								if (u != null) continue;
								e.removeEventListener(r, d[n], t), d[n] = null;
							}
							if (y) mr(r, e, u), hr([r]);
							else if (u != null) {
								function i(e) {
									d[a].call(this, e);
								}
								d[n] = fr(r, e, i, t);
							}
						} else if (a === "style") Y(e, a, u);
						else if (a === "autofocus") at(e, !!u);
						else if (!c && (a === "__value" || a === "value" && u != null)) e.__value = u, (_ == null || e.value !== u || u === 0 && e.nodeName === Ri) && (e.value = u);
						else if (a === "selected" && f) Ci(e, u);
						else {
							var b = a;
							l || (b = Nr(b));
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
							} else x || (c || typeof u != "string") && h.has(b) ? (e[b] = u, b in s && (s[b] = t)) : typeof u != "function" && Y(e, b, u, o);
						}
					}
				}
			}
		}
	} finally {
		u && E(!0);
	}
	return d;
}
function Vi(e, t, n = [], r = [], i = [], a, o = !1, s = !1) {
	lt(i, n, r, (n) => {
		var r = void 0, i = {}, c = e.nodeName === Li, l = !1;
		if (On(() => {
			var u = t(...n.map(J)), d = Bi(e, r, u, a, o, s);
			if (l && c) {
				var f = e;
				"defaultValue" in u && wi(f, u.defaultValue), "value" in u && Ei(f, u.value);
			}
			for (let e of Object.getOwnPropertySymbols(i)) u[e] || B(i[e]);
			for (let t of Object.getOwnPropertySymbols(u)) {
				var p = u[t];
				t.description !== "@attach" || r && p === r[t] || (i[t] && B(i[t]), i[t] = z(() => di(e, () => p))), d[t] = p;
			}
			r = d;
		}), c) {
			var u = e;
			Cn(() => {
				var e = r;
				"defaultValue" in e && wi(u, e.defaultValue), Ei(u, e.value, !0), Di(u);
			});
		}
		l = !0;
	});
}
function Hi(e) {
	return e[he] ??= {
		[Mi]: e.nodeName.includes("-"),
		[Ni]: e.namespaceURI === n
	};
}
var Ui = /* @__PURE__ */ new Map();
function Wi(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Ui.get(t);
	if (n) return n;
	Ui.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = u(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = p(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Gi = !1;
function Ki(e) {
	var t = Gi;
	try {
		return Gi = !1, [e(), Gi];
	} finally {
		Gi = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
var qi = {
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
function X(e, t, n) {
	return new Proxy({
		props: e,
		exclude: t
	}, qi);
}
var Ji = {
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
function Z(...e) {
	return new Proxy({ props: e }, Ji);
}
function Q(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ pt(r), J(u)) : (c && (c = !1, s = o ? or(r) : r), s);
	let f;
	if (a) {
		var p = de in e || pe in e;
		f = l(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Ki(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Ve(t), f(m)));
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
			let n = t ? J(y) : i && a ? Yt(e) : e;
			return I(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return Bn && v || b.f & 16384 ? y.v : J(y);
	});
}
// kit/plugin.svelte.js
function Yi(e, t) {
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
		#e = /* @__PURE__ */ F(Yt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return J(this.#e);
		}
		set props(e) {
			I(this.#e, e, !0);
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
			this.app ||= Ur(t, {
				target: this,
				props: this.props
			}), i || (this.style.visibility = "hidden", e.then(() => {
				this.style.visibility = "";
			}));
		}
		disconnectedCallback() {
			this.app && qr(this.app), this.app = null;
		}
	});
}
function Xi(e) {
	let t = Yt({ ...e()?.params?.() ?? {} });
	return bn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
function Zi(e, t) {
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
var Qi = {
	xmlns: "http://www.w3.org/2000/svg",
	width: 24,
	height: 24,
	viewBox: "0 0 24 24",
	fill: "none",
	stroke: "currentColor",
	"stroke-width": 2,
	"stroke-linecap": "round",
	"stroke-linejoin": "round"
}, $i = (...e) => e.filter((e, t, n) => !!e && e.trim() !== "" && n.indexOf(e) === t).join(" ").trim();
// node_modules/@lucide/svelte/dist/utils/buildLucideIconNode.js
function ea(e) {
	return e != null;
}
function ta(e, t = {}) {
	let n = t.attributeNames ?? {}, r = (e) => n[e] ?? e, i = e.size ?? e.width ?? Qi.width, a = e.size ?? e.height ?? Qi.height, o = e.aliases?.filter((e) => typeof e == "string" && e.trim() !== "").map((e) => `lucide-${e}`) ?? [], s = [...e.name ? [`lucide-${e.name}`] : [], ...o], c = t.className?.split(" ").filter(Boolean) ?? [], l = t.includeDefaultClasses === !1 ? $i(...c) : $i("lucide", ...s, ...c), u = t.absoluteStrokeWidth ? Number(t.strokeWidth ?? Qi["stroke-width"]) * Number(e.size ?? e.width ?? Qi.width) / Number(t.size ?? t.width ?? Qi.width) : t.strokeWidth ?? Qi["stroke-width"];
	return [
		"svg",
		{
			...Object.entries(Qi).reduce((e, [t, n]) => (e[r(t)] = n, e), {}),
			..."color" in t && t.color && { [r("stroke")]: t.color },
			..."size" in t && ea(t.size) && {
				[r("width")]: t.size,
				[r("height")]: t.size
			},
			..."width" in t && ea(t.width) && { [r("width")]: t.width },
			..."height" in t && ea(t.height) && { [r("height")]: t.height },
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
var na = (e) => {
	for (let t in e) if (t.startsWith("aria-") || t === "role" || t === "title") return !0;
	return !1;
}, ra = Symbol("lucide-context"), ia = () => Ye(ra), aa = /* @__PURE__ */ new Set([
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
]), oa = /* @__PURE__ */ Tr("<svg><!><!></svg>");
function $(e, t) {
	Xe(t, !0);
	let n = ia() ?? {}, r = Q(t, "color", 19, () => n.color ?? "currentColor"), i = Q(t, "size", 19, () => n.size ?? 24), a = Q(t, "width", 19, i), o = Q(t, "height", 19, i), s = Q(t, "strokeWidth", 19, () => n.strokeWidth ?? 2), c = Q(t, "absoluteStrokeWidth", 19, () => n.absoluteStrokeWidth ?? !1), l = Q(t, "nonScalingStroke", 19, () => n.nonScalingStroke ?? !1), u = Q(t, "iconNode", 19, () => []), d = Q(t, "icon", 19, () => ({
		node: u(),
		aliases: [],
		size: 24
	})), f = /* @__PURE__ */ X(t, aa), p = /* @__PURE__ */ M(() => !!t.children || na(f)), m = /* @__PURE__ */ M(() => ta(d(), {
		color: r(),
		width: a(),
		height: o(),
		strokeWidth: s(),
		absoluteStrokeWidth: c(),
		nonScalingStroke: l(),
		className: $i("lucide-icon", n.class),
		hasA11yProp: J(p),
		attributes: f
	})), h = /* @__PURE__ */ M(() => b(J(m), 3)), _ = /* @__PURE__ */ M(() => J(h)[1]), v = /* @__PURE__ */ M(() => y(J(h)[2], () => [], !0)), x = /* @__PURE__ */ M(() => ({
		...J(_),
		class: [...J(_).class.split(" "), t.class]
	}));
	var S = oa();
	Vi(S, () => ({ ...J(x) }));
	var C = an(S);
	ni(C, 17, () => J(v), Qr, (e, t) => {
		var n = /* @__PURE__ */ M(() => b(J(t), 2));
		let r = () => J(n)[0], i = () => J(n)[1];
		var a = Er();
		li(on(a), r, !0, (e, t) => {
			Vi(e, () => ({ ...i() }));
		}), Dr(e, a);
	}), Yr(cn(C), () => t.children ?? g), De(S), Dr(e, S), Ze();
}
// node_modules/@lucide/svelte/dist/icons/sun.svelte
var sa = {
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
sa.node;
var ca = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function la(e, t) {
	let n = /* @__PURE__ */ X(t, ca);
	$(e, Z(() => n, { get icon() {
		return sa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/moon.svelte
var ua = {
	name: "moon",
	size: 24,
	node: [["path", { d: "M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401" }]]
};
ua.node;
var da = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function fa(e, t) {
	let n = /* @__PURE__ */ X(t, da);
	$(e, Z(() => n, { get icon() {
		return ua;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/inbox.svelte
var pa = {
	name: "inbox",
	size: 24,
	node: [["polyline", { points: "22 12 16 12 14 15 10 15 8 12 2 12" }], ["path", { d: "M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" }]]
};
pa.node;
var ma = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ha(e, t) {
	let n = /* @__PURE__ */ X(t, ma);
	$(e, Z(() => n, { get icon() {
		return pa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/message-square.svelte
var ga = {
	name: "message-square",
	size: 24,
	node: [["path", { d: "M22 17a2 2 0 0 1-2 2H6.828a2 2 0 0 0-1.414.586l-2.202 2.202A.71.71 0 0 1 2 21.286V5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2z" }]]
};
ga.node;
var _a = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function va(e, t) {
	let n = /* @__PURE__ */ X(t, _a);
	$(e, Z(() => n, { get icon() {
		return ga;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/list-checks.svelte
var ya = {
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
ya.node;
var ba = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function xa(e, t) {
	let n = /* @__PURE__ */ X(t, ba);
	$(e, Z(() => n, { get icon() {
		return ya;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/activity.svelte
var Sa = {
	name: "activity",
	size: 24,
	node: [["path", { d: "M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" }]]
};
Sa.node;
var Ca = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function wa(e, t) {
	let n = /* @__PURE__ */ X(t, Ca);
	$(e, Z(() => n, { get icon() {
		return Sa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/terminal.svelte
var Ta = {
	name: "terminal",
	size: 24,
	node: [["path", { d: "M12 19h8" }], ["path", { d: "m4 17 6-6-6-6" }]]
};
Ta.node;
var Ea = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Da(e, t) {
	let n = /* @__PURE__ */ X(t, Ea);
	$(e, Z(() => n, { get icon() {
		return Ta;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/boxes.svelte
var Oa = {
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
Oa.node;
var ka = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Aa(e, t) {
	let n = /* @__PURE__ */ X(t, ka);
	$(e, Z(() => n, { get icon() {
		return Oa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/rocket.svelte
var ja = {
	name: "rocket",
	size: 24,
	node: [
		["path", { d: "M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5" }],
		["path", { d: "M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09" }],
		["path", { d: "M9 12a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.4 22.4 0 0 1-4 2z" }],
		["path", { d: "M9 12H4s.55-3.03 2-4c1.62-1.08 5 .05 5 .05" }]
	]
};
ja.node;
var Ma = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Na(e, t) {
	let n = /* @__PURE__ */ X(t, Ma);
	$(e, Z(() => n, { get icon() {
		return ja;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/puzzle.svelte
var Pa = {
	name: "puzzle",
	size: 24,
	node: [["path", { d: "M15.39 4.39a1 1 0 0 0 1.68-.474 2.5 2.5 0 1 1 3.014 3.015 1 1 0 0 0-.474 1.68l1.683 1.682a2.414 2.414 0 0 1 0 3.414L19.61 15.39a1 1 0 0 1-1.68-.474 2.5 2.5 0 1 0-3.014 3.015 1 1 0 0 1 .474 1.68l-1.683 1.682a2.414 2.414 0 0 1-3.414 0L8.61 19.61a1 1 0 0 0-1.68.474 2.5 2.5 0 1 1-3.014-3.015 1 1 0 0 0 .474-1.68l-1.683-1.682a2.414 2.414 0 0 1 0-3.414L4.39 8.61a1 1 0 0 1 1.68.474 2.5 2.5 0 1 0 3.014-3.015 1 1 0 0 1-.474-1.68l1.683-1.682a2.414 2.414 0 0 1 3.414 0z" }]]
};
Pa.node;
var Fa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ia(e, t) {
	let n = /* @__PURE__ */ X(t, Fa);
	$(e, Z(() => n, { get icon() {
		return Pa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/layout-dashboard.svelte
var La = {
	name: "layout-dashboard",
	size: 24,
	node: [
		["rect", {
			width: "7",
			height: "9",
			x: "3",
			y: "3",
			rx: "1"
		}],
		["rect", {
			width: "7",
			height: "5",
			x: "14",
			y: "3",
			rx: "1"
		}],
		["rect", {
			width: "7",
			height: "9",
			x: "14",
			y: "12",
			rx: "1"
		}],
		["rect", {
			width: "7",
			height: "5",
			x: "3",
			y: "16",
			rx: "1"
		}]
	]
};
La.node;
var Ra = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function za(e, t) {
	let n = /* @__PURE__ */ X(t, Ra);
	$(e, Z(() => n, { get icon() {
		return La;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/settings.svelte
var Ba = {
	name: "settings",
	size: 24,
	node: [["path", { d: "M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915" }], ["circle", {
		cx: "12",
		cy: "12",
		r: "3"
	}]]
};
Ba.node;
var Va = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ha(e, t) {
	let n = /* @__PURE__ */ X(t, Va);
	$(e, Z(() => n, { get icon() {
		return Ba;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/bell.svelte
var Ua = {
	name: "bell",
	size: 24,
	node: [["path", { d: "M10.268 21a2 2 0 0 0 3.464 0" }], ["path", { d: "M3.262 15.326A1 1 0 0 0 4 17h16a1 1 0 0 0 .74-1.673C19.41 13.956 18 12.499 18 8A6 6 0 0 0 6 8c0 4.499-1.411 5.956-2.738 7.326" }]]
};
Ua.node;
var Wa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ga(e, t) {
	let n = /* @__PURE__ */ X(t, Wa);
	$(e, Z(() => n, { get icon() {
		return Ua;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/file-text.svelte
var Ka = {
	name: "file-text",
	size: 24,
	node: [
		["path", { d: "M6 22a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h8a2.4 2.4 0 0 1 1.704.706l3.588 3.588A2.4 2.4 0 0 1 20 8v12a2 2 0 0 1-2 2z" }],
		["path", { d: "M14 2v5a1 1 0 0 0 1 1h5" }],
		["path", { d: "M10 9H8" }],
		["path", { d: "M16 13H8" }],
		["path", { d: "M16 17H8" }]
	]
};
Ka.node;
var qa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Ja(e, t) {
	let n = /* @__PURE__ */ X(t, qa);
	$(e, Z(() => n, { get icon() {
		return Ka;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/git-pull-request.svelte
var Ya = {
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
Ya.node;
var Xa = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function Za(e, t) {
	let n = /* @__PURE__ */ X(t, Xa);
	$(e, Z(() => n, { get icon() {
		return Ya;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/chart-bar.svelte
var Qa = {
	name: "chart-bar",
	size: 24,
	node: [
		["path", { d: "M3 3v16a2 2 0 0 0 2 2h16" }],
		["path", { d: "M7 16h8" }],
		["path", { d: "M7 11h12" }],
		["path", { d: "M7 6h3" }]
	],
	aliases: ["bar-chart-horizontal"]
};
Qa.node;
var $a = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function eo(e, t) {
	let n = /* @__PURE__ */ X(t, $a);
	$(e, Z(() => n, { get icon() {
		return Qa;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/search.svelte
var to = {
	name: "search",
	size: 24,
	node: [["path", { d: "m21 21-4.34-4.34" }], ["circle", {
		cx: "11",
		cy: "11",
		r: "8"
	}]]
};
to.node;
var no = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function ro(e, t) {
	let n = /* @__PURE__ */ X(t, no);
	$(e, Z(() => n, { get icon() {
		return to;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/server.svelte
var io = {
	name: "server",
	size: 24,
	node: [
		["rect", {
			width: "20",
			height: "8",
			x: "2",
			y: "2",
			rx: "2",
			ry: "2"
		}],
		["rect", {
			width: "20",
			height: "8",
			x: "2",
			y: "14",
			rx: "2",
			ry: "2"
		}],
		["line", {
			x1: "6",
			x2: "6.01",
			y1: "6",
			y2: "6"
		}],
		["line", {
			x1: "6",
			x2: "6.01",
			y1: "18",
			y2: "18"
		}]
	]
};
io.node;
var ao = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function oo(e, t) {
	let n = /* @__PURE__ */ X(t, ao);
	$(e, Z(() => n, { get icon() {
		return io;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/users.svelte
var so = {
	name: "users",
	size: 24,
	node: [
		["path", { d: "M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" }],
		["path", { d: "M16 3.128a4 4 0 0 1 0 7.744" }],
		["path", { d: "M22 21v-2a4 4 0 0 0-3-3.87" }],
		["circle", {
			cx: "9",
			cy: "7",
			r: "4"
		}]
	]
};
so.node;
var co = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function lo(e, t) {
	let n = /* @__PURE__ */ X(t, co);
	$(e, Z(() => n, { get icon() {
		return so;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/database.svelte
var uo = {
	name: "database",
	size: 24,
	node: [
		["ellipse", {
			cx: "12",
			cy: "5",
			rx: "9",
			ry: "3"
		}],
		["path", { d: "M3 5V19A9 3 0 0 0 21 19V5" }],
		["path", { d: "M3 12A9 3 0 0 0 21 12" }]
	]
};
uo.node;
var fo = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function po(e, t) {
	let n = /* @__PURE__ */ X(t, fo);
	$(e, Z(() => n, { get icon() {
		return uo;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/shield.svelte
var mo = {
	name: "shield",
	size: 24,
	node: [["path", { d: "M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z" }]]
};
mo.node;
var ho = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function go(e, t) {
	let n = /* @__PURE__ */ X(t, ho);
	$(e, Z(() => n, { get icon() {
		return mo;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/clock.svelte
var _o = {
	name: "clock",
	size: 24,
	node: [["circle", {
		cx: "12",
		cy: "12",
		r: "10"
	}], ["path", { d: "M12 6v6l4 2" }]]
};
_o.node;
var vo = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function yo(e, t) {
	let n = /* @__PURE__ */ X(t, vo);
	$(e, Z(() => n, { get icon() {
		return _o;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/zap.svelte
var bo = {
	name: "zap",
	size: 24,
	node: [["path", { d: "M15.914 4a1.5 1.5 0 00-2.474-1.561l-9 9A1.5 1.5 0 005.5 14h4.002a.5.5 0 01.471.666L8.086 20a1.5 1.5 0 002.475 1.56l9-9A1.5 1.5 0 0018.5 10h-3.997a.5.5 0 01-.472-.667z" }]]
};
bo.node;
var xo = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function So(e, t) {
	let n = /* @__PURE__ */ X(t, xo);
	$(e, Z(() => n, { get icon() {
		return bo;
	} }));
}
// node_modules/@lucide/svelte/dist/icons/house.svelte
var Co = {
	name: "house",
	size: 24,
	node: [["path", { d: "M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8" }], ["path", { d: "M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" }]],
	aliases: ["home"]
};
Co.node;
var wo = /* @__PURE__ */ new Set([
	"$$slots",
	"$$events",
	"$$legacy"
]);
function To(e, t) {
	let n = /* @__PURE__ */ X(t, wo);
	$(e, Z(() => n, { get icon() {
		return Co;
	} }));
}
// plugins/nav/icons.js
var Eo = {
	inbox: ha,
	"message-square": va,
	"list-checks": xa,
	activity: wa,
	terminal: Da,
	boxes: Aa,
	rocket: Na,
	puzzle: Ia,
	"layout-dashboard": za,
	settings: Ha,
	bell: Ga,
	"file-text": Ja,
	"git-pull-request": Za,
	"chart-bar": eo,
	search: ro,
	server: oo,
	users: lo,
	database: po,
	shield: go,
	clock: yo,
	zap: So,
	house: To
}, Do = (e) => Eo[e ?? ""] ?? Ia, Oo = /* @__PURE__ */ Cr("<div class=\"user faint svelte-eck48c\" title=\"signed in through the proxy\"> </div>"), ko = /* @__PURE__ */ Cr("<span class=\"tabdot svelte-eck48c\"></span>"), Ao = /* @__PURE__ */ Cr("<span class=\"badge svelte-eck48c\"> </span>"), jo = /* @__PURE__ */ Cr("<a><span class=\"ico svelte-eck48c\"><!><!></span> <span class=\"lbl svelte-eck48c\"> </span> <!></a>"), Mo = /* @__PURE__ */ Cr("<div class=\"hdr plugin-host svelte-eck48c\"></div>"), No = /* @__PURE__ */ Cr("<nav aria-label=\"views\"><div class=\"links svelte-eck48c\"><!> <!> <!></div> <div class=\"foot svelte-eck48c\"><span class=\"live svelte-eck48c\"><span></span> <span class=\"faint\"> </span></span> <span class=\"tools svelte-eck48c\"><button class=\"ghost icon svelte-eck48c\" aria-label=\"toggle theme\" title=\"toggle theme\"><!></button></span></div></nav>");
function Po(e, t) {
	Xe(t, !0);
	let n = Q(t, "api", 7), r = Q(t, "context", 19, () => ({})), i = Xi(() => n()), a = /* @__PURE__ */ M(() => r().dock || "left"), o = /* @__PURE__ */ M(() => J(a) === "top" || J(a) === "bottom"), s = /* @__PURE__ */ M(() => J(o) && !!r().narrow), c = /* @__PURE__ */ F(""), l = /* @__PURE__ */ F(!1), u = /* @__PURE__ */ F(Yt({})), d = /* @__PURE__ */ F(!1);
	Xr(() => {
		I(d, n().theme === "light"), I(c, n().user, !0), J(c) || n().get("api/whoami").then((e) => I(c, e.user || "", !0)).catch(() => {});
		let e = n().live((e) => I(l, e, !0)), t = n().badges((e) => I(u, e, !0));
		return () => {
			e(), t();
		};
	});
	function f() {
		I(d, !J(d)), n().theme = J(d) ? "light" : "dark";
	}
	let p = /* @__PURE__ */ M(() => n().views()), m = /* @__PURE__ */ M(() => n().contributions("header")), h = /* @__PURE__ */ M(() => i.view || "attention"), g = (e) => n().go(e), _ = 0;
	function v(e) {
		let t = e.target;
		if (!(e.metaKey || e.ctrlKey || e.altKey || t.closest("input, textarea, select, [contenteditable]"))) {
			if (_ && Date.now() - _ < 1200) {
				_ = 0;
				let t = J(p).find((t) => t.key && t.key === e.key);
				t && (e.preventDefault(), g({
					view: t.id,
					ticket: ""
				}));
			} else e.key === "g" && (_ = Date.now());
		}
	}
	var y = No();
	pr("keydown", Qt, v);
	let b;
	var x = an(y), S = an(x), C = (e) => {
		var t = Oo(), n = sn(t, !0);
		En(() => Hr(n, J(c))), Dr(e, t);
	};
	Zr(S, (e) => {
		J(c) && e(C);
	});
	var w = cn(S, 2);
	ni(w, 17, () => J(p), (e) => e.id, (e, t) => {
		let n = /* @__PURE__ */ M(() => Do(J(t).icon));
		var r = jo();
		let i;
		var a = an(r), o = an(a);
		ci(o, () => J(n), (e, t) => {
			t(e, { size: 14 });
		});
		var s = cn(o), c = (e) => {
			Dr(e, ko());
		};
		Zr(s, (e) => {
			(J(u)[J(t).id] ?? 0) > 0 && e(c);
		}), De(a);
		var l = cn(a, 2), d = sn(l, !0), f = cn(l, 2), p = (e) => {
			var n = Ao(), r = sn(n, !0);
			En(() => Hr(r, J(u)[J(t).id])), Dr(e, n);
		};
		Zr(f, (e) => {
			(J(u)[J(t).id] ?? 0) > 0 && e(p);
		}), De(r), En(() => {
			Y(r, "href", "?view=" + J(t).id), Y(r, "data-view", J(t).id), Y(r, "title", "Go to " + J(t).title + (J(t).key ? " (g " + J(t).key + ")" : "")), i = bi(r, 1, "svelte-eck48c", null, i, { active: J(h) === J(t).id }), Hr(d, J(t).title);
		}), mr("click", r, (e) => {
			e.preventDefault(), g({
				view: J(t).id,
				ticket: ""
			});
		}), Dr(e, r);
	}), ni(cn(w, 2), 17, () => J(m), (e) => e.plugin + e.element, (e, t) => {
		var r = Mo();
		ui(r, (e, t) => Zi?.(e, t), () => ({
			api: n(),
			slot: "header",
			target: {
				plugin: J(t).plugin,
				element: J(t).element
			}
		})), En(() => Y(r, "data-plugin", J(t).plugin)), Dr(e, r);
	}), De(x);
	var ee = cn(x, 2), te = an(ee), ne = an(te);
	let re;
	var ie = sn(cn(ne, 2), !0);
	De(te);
	var ae = cn(te, 2), oe = an(ae), se = an(oe), ce = (e) => {
		fa(e, { size: 14 });
	}, le = (e) => {
		la(e, { size: 14 });
	};
	Zr(se, (e) => {
		J(d) ? e(ce) : e(le, -1);
	}), De(oe), De(ae), De(ee), De(y), En(() => {
		b = bi(y, 1, `nav ${J(a) ?? ""}`, "svelte-eck48c", b, {
			bar: J(o),
			tabs: J(s)
		}), Y(y, "data-dock", J(a)), Y(te, "title", J(l) ? "live" : "disconnected"), re = bi(ne, 1, "dot", null, re, {
			ok: J(l),
			bad: !J(l)
		}), Hr(ie, J(l) ? "live" : "offline");
	}), mr("click", oe, f), Dr(e, y), Ze();
}
// plugins/nav/index.js
hr(["click"]), Yi("desk-nav", Po);
