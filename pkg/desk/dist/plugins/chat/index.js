// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = "http://www.w3.org/2000/svg", i = "http://www.w3.org/1998/Math/MathML", a = Array.isArray, o = Array.prototype.indexOf, s = Array.prototype.includes, c = Array.from, l = Object.defineProperty, u = Object.getOwnPropertyDescriptor, d = Object.getOwnPropertyDescriptors, f = Object.prototype, p = Array.prototype, m = Object.getPrototypeOf, h = Object.isExtensible, g = () => {};
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
function y(e, t) {
	if (Array.isArray(e)) return e;
	if (t === void 0 || !(Symbol.iterator in e)) return Array.from(e);
	let n = [];
	for (let r of e) if (n.push(r), n.length === t) break;
	return n;
}
var b = 1024, x = 2048, S = 4096, C = 8192, w = 16384, ee = 32768, te = 1 << 25, T = 65536, ne = 1 << 19, re = 1 << 20, ie = 1 << 25, ae = 1 << 21, oe = 1 << 22, se = 1 << 23, ce = Symbol("$state"), le = Symbol("component"), ue = Symbol("legacy props"), de = Symbol(""), fe = Symbol("attributes"), pe = Symbol("class"), me = Symbol("style"), he = Symbol("text"), E = Symbol("form reset"), ge = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), _e = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function ve() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ye(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function be() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var D = !1;
function xe(e) {
	D = e;
}
var O;
function Se(t) {
	if (t === null) throw ye(), e;
	return O = t;
}
function k() {
	return Se(/* @__PURE__ */ nn(O));
}
function A(t) {
	if (D) {
		if (/* @__PURE__ */ nn(O) !== null) throw ye(), e;
		O = t;
	}
}
function j(e = 1) {
	if (D) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ nn(n);
		O = n;
	}
}
function Ce(e = !0) {
	for (var t = 0, n = O;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ nn(n);
		e && n.remove(), n = i;
	}
}
function we(t) {
	if (!t || t.nodeType !== 8) throw ye(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Te(e) {
	return e === this.v;
}
function Ee(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function De(e) {
	return !Ee(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function Oe() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function ke(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Ae(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function je() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Me(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Ne() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Pe(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Le() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Re() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/shared/clone.js
var ze = [];
function Be(e, t = !1, n = !1) {
	return Ve(e, /* @__PURE__ */ new Map(), "", ze, null, n);
}
function Ve(e, t, n, r, i = null, o = !1) {
	if (typeof e == "object" && e) {
		var s = t.get(e);
		if (s !== void 0) return s;
		if (e instanceof Map) return new Map(e);
		if (e instanceof Set) return new Set(e);
		if (a(e)) {
			var c = Array(e.length);
			t.set(e, c), i !== null && t.set(i, c);
			for (var l = 0; l < e.length; l += 1) {
				var u = e[l];
				l in e && (c[l] = Ve(u, t, n, r, null, o));
			}
			return c;
		}
		if (m(e) === f) {
			c = {}, t.set(e, c), i !== null && t.set(i, c);
			for (var d of Object.keys(e)) c[d] = Ve(e[d], t, n, r, null, o);
			return c;
		}
		if (e instanceof Date) return e.getTime(), structuredClone(e);
		if (typeof e.toJSON == "function" && !o) return Ve(e.toJSON(), t, n, r, e);
	}
	if (e instanceof EventTarget) return e;
	try {
		return structuredClone(e);
	} catch {
		return e;
	}
}
// node_modules/svelte/src/internal/client/context.js
var He = null;
function Ue(e) {
	He = e;
}
function We(e, t = !1, n) {
	He = {
		p: He,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: H,
		l: null
	};
}
function Ge(e) {
	var t = He, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) _n(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, He = t.p, Ke(e);
}
function Ke(e = {}) {
	return l(e, le, { value: !0 }), e;
}
function qe() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Je = [];
function Ye() {
	var e = Je;
	Je = [], _(e);
}
function Xe(e) {
	if (Je.length === 0 && !Ct) {
		var t = Je;
		queueMicrotask(() => {
			t === Je && Ye();
		});
	}
	Je.push(e);
}
function Ze() {
	for (; Je.length > 0;) Ye();
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Qe = ~(x | S | b);
function M(e, t) {
	e.f = e.f & Qe | t;
}
function $e(e) {
	e.f & 512 || e.deps === null ? M(e, b) : M(e, S);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function et(e, t) {
	e.f & 2048 ? t.set(e, x) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, S), M(e, b);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function tt(e) {
	D && /* @__PURE__ */ tn(e) !== null && an(e);
}
var nt = !1;
function rt() {
	nt || (nt = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[E]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function it(e) {
	var t = V, n = H;
	zn(null), Bn(null);
	try {
		return e();
	} finally {
		zn(t), Bn(n);
	}
}
function at(e, t, n, r = n) {
	e.addEventListener(t, () => it(n));
	let i = e[E];
	e[E] = i ? () => {
		i(), r(!0);
	} : () => r(!0), rt();
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function ot(e, t, n, r) {
	let i = qe() ? ut : N;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = H, c = st(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				un(e, s);
			}
			ct();
		}
	}
	var d = lt();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ ft(e))).then(u).catch((e) => un(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), ct());
	}) : f();
}
function st() {
	var e = H, t = V, n = He, r = P;
	return function(i = !0) {
		Bn(e), zn(t), Ue(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function ct(e = !0) {
	Bn(null), zn(null), Ue(null), e && P?.deactivate();
}
function lt() {
	var e = H, t = e.b, n = P, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function ut(e) {
	var n = 2 | x;
	return H !== null && (H.f |= ne), {
		ctx: He,
		deps: null,
		effects: null,
		equals: Te,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: H,
		ac: null
	};
}
var dt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function ft(e, n, r) {
	let i = H;
	i === null && Oe();
	var a = void 0, o = Bt(t), s = !V, c = /* @__PURE__ */ new Set();
	return bn(() => {
		var t = H, n = v();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ge && n.reject(e);
			}).finally(ct);
		} catch (e) {
			n.reject(e), ct();
		}
		var r = P;
		if (s) {
			if (t.f & 32768) var l = lt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(dt);
			else for (let e of c.values()) e.reject(dt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== dt && (r.activate(), t ? (o.f |= se, Wt(o, t)) : (o.f & 8388608 && (o.f ^= se), Wt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), hn(() => {
		for (let e of c) e.reject(dt);
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
function pt(e) {
	let t = /* @__PURE__ */ ut(e);
	return Hn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function N(e) {
	let t = /* @__PURE__ */ ut(e);
	return t.equals = De, t;
}
function mt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) Dn(t[n]);
	}
}
function ht(e) {
	var n, r = H, i = e.parent;
	if (!In && i !== null && e.v !== t && i.f & 24576) return ve(), e.v;
	Bn(i);
	try {
		mt(e), n = er(e);
	} finally {
		Bn(r);
	}
	return n;
}
function gt(e) {
	var t = ht(e);
	!e.equals(t) && (e.wv = Zn(), (!P?.is_fork || e.deps === null) && (P === null ? e.v = t : (P.capture(e, t, !0), bt?.capture(e, t, !0)), e.deps === null)) ? M(e, b) : In || (xt === null ? $e(e) : (mn() || P?.is_fork) && xt.set(e, t));
}
function _t(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && it(() => {
		t.ac.abort(ge), t.ac = null;
	}), t.fn !== null && (t.teardown = g), rr(t, 0), Tn(t));
}
function vt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && ir(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var yt = null, P = null, bt = null, xt = null, St = null, Ct = !1, wt = !1, Tt = null, Et = null, Dt = 0, Ot = 1, kt = class e {
	id = Ot++;
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
		yt === null ? yt = this : (yt.next = this, this.prev = yt), yt = this;
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
			for (var r of n.d) M(r, x), t(r);
			for (r of n.m) M(r, S), t(r);
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
					t.f ^= b;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) M(e, t), this.schedule(e);
		this.apply();
		for (var t = Tt = [], n = [], r = Et = []; this.#a.length > 0;) {
			Dt++ > 1e3 && (this.#b(), jt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw It(e), this.#d() || this.discard(), t;
			}
		}
		if (P = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (Tt = null, Et = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Ft(e, t);
			r.length > 0 && P.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), bt = this, Nt(n), Nt(t), bt = null, this.#i?.resolve();
			var o = P;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Rt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= b;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= b : i & 4 ? t.push(r) : Qn(r) && (i & 16 && this.add_dirty_reaction(r, S), ir(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, x), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), P = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) et(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), xt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		P = this;
	}
	deactivate() {
		P = null, xt = null;
	}
	flush() {
		try {
			wt = !0, P = this, this.#m();
		} finally {
			Dt = 0, St = null, Tt = null, Et = null, wt = !1, P = null, xt = null, Rt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(dt);
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
		this.#u || (this.#u = !0, Xe(() => {
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
		if (P === null) {
			let t = P = new e();
			!wt && !Ct && Xe(() => {
				t.started || t.flush();
			});
		}
		return P;
	}
	apply() {
		xt = null;
	}
	schedule(e) {
		St = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? yt = e : t.prev = e, this.linked = !1;
		}
	}
};
function At(e) {
	var t = Ct, n = bt;
	bt = null, Ct = !0;
	try {
		var r;
		for (e && (At(), r = e());;) {
			if (Ze(), P === null) return r;
			P.flush();
		}
	} finally {
		Ct = t, bt = n;
	}
}
function jt() {
	try {
		Ne();
	} catch (e) {
		un(e, St);
	}
}
var Mt = null;
function Nt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Qn(r) && (Mt = /* @__PURE__ */ new Set(), ir(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && kn(r), Mt?.size > 0)) {
				Rt.clear();
				for (let e of Mt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Mt.has(n) && (Mt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || ir(n);
					}
				}
				Mt.clear();
			}
		}
		Mt = null;
	}
}
function Pt(e) {
	P.schedule(e);
}
function Ft(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, b);
		for (var n = e.first; n !== null;) Ft(n, t), n = n.next;
	}
}
function It(e) {
	M(e, b);
	for (var t = e.first; t !== null;) It(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Lt = /* @__PURE__ */ new Set(), Rt = /* @__PURE__ */ new Map(), zt = !1;
function Bt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Te,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function F(e, t) {
	let n = Bt(e, t);
	return Hn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Vt(e, t = !1, n = !0) {
	let r = Bt(e);
	return t || (r.equals = De), r;
}
function I(e, t, n = !1) {
	return V !== null && (!Rn || V.f & 131072) && qe() && V.f & 4325394 && (Vn === null || !Vn.has(e)) && Le(), Wt(e, n ? Jt(t) : t, Et);
}
var Ht = null, Ut = 0;
function Wt(e, t, n = null) {
	if (!e.equals(t)) {
		In ? Rt.set(e, t) : Rt.has(e) || Rt.set(e, e.v);
		var r = kt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ht(t), xt === null && $e(t);
		}
		e.wv = Zn(), Ht = null, Ut = 0, qt(e, x, n), Ht = null, qe() && H !== null && H.f & 1024 && !(H.f & 96) && (Gn === null ? Kn([e]) : Gn.push(e)), !r.is_fork && Lt.size > 0 && !zt && Gt();
	}
	return t;
}
function Gt() {
	zt = !1;
	for (let e of Lt) {
		e.f & 1024 && M(e, S);
		let t;
		try {
			t = Qn(e);
		} catch {
			t = !0;
		}
		t && ir(e);
	}
	Lt.clear();
}
function Kt(e) {
	I(e, e.v + 1);
}
function qt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = qe(), a = r.length;
		if (Ut += a, Ut > 1e5 && Ht === null && (Ht = /* @__PURE__ */ new Set()), Ht !== null) {
			if (Ht.has(e)) return;
			Ht.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== H) {
				var l = (c & x) === 0;
				if (l && M(s, t), c & 131072) Lt.add(s);
				else if (c & 2) {
					var u = s;
					xt?.delete(u), qt(u, S, n);
				} else if (l) {
					var d = s;
					c & 16 && Mt !== null && Mt.add(d), n === null ? Pt(d) : n.push(d);
				}
			}
		}
	}
}
function Jt(e) {
	if (typeof e != "object" || !e || ce in e || le in e) return e;
	let n = m(e);
	if (n !== f && n !== p) return e;
	var r = /* @__PURE__ */ new Map(), i = a(e), o = /* @__PURE__ */ F(0), s = null, c = Yn, l = (e) => {
		if (Yn === c) return e();
		var t = V, n = Yn;
		zn(null), Xn(c);
		var r = e();
		return zn(t), Xn(n), r;
	};
	return i && r.set("length", /* @__PURE__ */ F(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Fe();
			var i = r.get(t);
			return i === void 0 ? l(() => {
				var e = /* @__PURE__ */ F(n.value, s);
				return r.set(t, e), e;
			}) : I(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = l(() => /* @__PURE__ */ F(t, s));
					r.set(n, e), Kt(o);
				}
			} else I(i, t), Kt(o);
			return !0;
		},
		get(n, i, a) {
			if (i === ce) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || u(n, i)?.writable) && (o = l(() => /* @__PURE__ */ F(Jt(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = U(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = U(a);
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
			if (n === ce) return !0;
			var i = r.get(n), a = i !== void 0 && i.v !== t || Reflect.has(e, n);
			return (i !== void 0 || (H !== null || V !== null) && (!a || u(e, n)?.writable)) && (i === void 0 && (i = l(() => /* @__PURE__ */ F(a ? Jt(e[n]) : t, s)), r.set(n, i)), U(i) === t) ? !1 : a;
		},
		set(e, n, a, c) {
			var d = r.get(n), f = n in e;
			if (i && n === "length") for (var p = a; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = l(() => /* @__PURE__ */ F(t, s)), r.set(p + "", m)) : I(m, t);
			}
			if (d === void 0) (!f || u(e, n)?.writable) && (d = l(() => /* @__PURE__ */ F(void 0, s)), I(d, Jt(a)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = l(() => Jt(a));
				I(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, a), !f) {
				if (i && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && I(_, v + 1);
				}
				Kt(o);
			}
			return !0;
		},
		ownKeys(e) {
			U(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = r.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [i, a] of r) a.v !== t && !(i in e) && n.push(i);
			return n;
		},
		setPrototypeOf() {
			Ie();
		}
	});
}
var Yt, Xt, Zt, Qt;
function $t() {
	if (Yt === void 0) {
		Yt = window, Xt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Zt = u(t, "firstChild").get, Qt = u(t, "nextSibling").get, h(e) && (e[pe] = void 0, e[fe] = null, e[me] = void 0, e.__e = void 0), h(n) && (n[he] = void 0);
	}
}
function en(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function tn(e) {
	return Zt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function nn(e) {
	return Qt.call(e);
}
function L(e, t) {
	if (!D) return /* @__PURE__ */ tn(e);
	var n = /* @__PURE__ */ tn(O);
	if (n === null) n = O.appendChild(en());
	else if (t && n.nodeType !== 3) {
		var r = en();
		return n?.before(r), Se(r), r;
	}
	return t && cn(n), Se(n), n;
}
function rn(e, t = !1) {
	if (!D) {
		var n = /* @__PURE__ */ tn(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ nn(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = en();
			return O?.before(r), Se(r), r;
		}
		cn(O);
	}
	return O;
}
function R(e, t = !1) {
	if (!D) return /* @__PURE__ */ tn(e);
	var n = L(e, t);
	return A(e), n;
}
function z(e, t = 1, n = !1) {
	let r = D ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ nn(r);
	if (!D) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = en();
			return r === null ? i?.after(a) : r.before(a), Se(a), a;
		}
		cn(r);
	}
	return Se(r), r;
}
function an(e) {
	e.textContent = "";
}
function on() {
	return !1;
}
function sn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function cn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function ln(e) {
	var t = H;
	if (t === null) return V.f |= se, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	un(e, t);
}
function un(t, n) {
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
function dn(e) {
	H === null && (V === null && Me(e), je()), In && Ae(e);
}
function fn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function pn(e, t) {
	var n = H;
	n !== null && n.f & 8192 && (e |= C);
	var r = {
		ctx: He,
		deps: null,
		nodes: null,
		f: e | x | 512,
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
	P?.register_created_effect(r);
	var i = r;
	if (e & 4) Tt === null ? kt.ensure().schedule(r) : Tt.push(r);
	else if (t !== null) {
		try {
			ir(r);
		} catch (e) {
			throw Dn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= T));
	}
	if (i !== null && (i.parent = n, n !== null && fn(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function mn() {
	return V !== null && !Rn;
}
function hn(e) {
	let t = pn(8, null);
	return M(t, b), t.teardown = e, t;
}
function gn(e) {
	dn("$effect");
	var t = H.f;
	if (!V && t & 32 && He !== null && !He.i) {
		var n = He;
		(n.e ??= []).push(e);
	} else return _n(e);
}
function _n(e) {
	return pn(4 | re, e);
}
function vn(e) {
	kt.ensure();
	let t = pn(64 | ne, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? An(t, () => {
			Dn(t), n(void 0);
		}) : (Dn(t), n(void 0));
	});
}
function yn(e) {
	return pn(4, e);
}
function bn(e) {
	return pn(oe | ne, e);
}
function xn(e, t = 0) {
	return pn(8 | t, e);
}
function B(e, t = [], n = [], r = []) {
	ot(r, t, n, (t) => {
		pn(8, () => {
			e(...t.map(U));
		});
	});
}
function Sn(e, t = 0) {
	return pn(16 | t, e);
}
function Cn(e) {
	return pn(32 | ne, e);
}
function wn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = In, r = V;
		Ln(!0), zn(null);
		try {
			t.call(null);
		} catch (t) {
			un(t, e.parent);
		} finally {
			Ln(n), zn(r);
		}
	}
}
function Tn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && it(() => {
			e.abort(ge);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : Dn(n, t), n = r;
	}
}
function En(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || Dn(t), t = n;
	}
}
function Dn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (On(e.nodes.start, e.nodes.end), n = !0), e.f |= te, Tn(e, t && !n), rr(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	wn(e), e.f ^= te, e.f |= w;
	var i = e.parent;
	i !== null && i.first !== null && kn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function On(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ nn(e);
		e.remove(), e = n;
	}
}
function kn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function An(e, t, n = !0) {
	var r = [];
	e.f |= 256, jn(e, r, !0);
	var i = () => {
		n && Dn(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function jn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= C;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				jn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Mn(e) {
	e.f &= -257, Nn(e, !0);
}
function Nn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= C, e.f & 1024 || (M(e, x), kt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Nn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Pn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ nn(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Fn = null, In = !1;
function Ln(e) {
	In = e;
}
var V = null, Rn = !1;
function zn(e) {
	V = e;
}
var H = null;
function Bn(e) {
	H = e;
}
var Vn = null;
function Hn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (Vn ??= /* @__PURE__ */ new Set()).add(e);
}
var Un = null, Wn = 0, Gn = null;
function Kn(e) {
	Gn = e;
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
			if (Qn(a) && gt(a), a.wv > e.wv) return !0;
		}
		t & 512 && xt === null && M(e, b);
	}
	return !1;
}
function $n(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Vn !== null && Vn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? $n(a, t, !1) : t === a && (n ? M(a, x) : a.f & 1024 && M(a, S), Pt(a));
	}
}
function er(e) {
	var t = Un, n = Wn, r = Gn, i = V, a = Vn, o = He, s = Rn, c = Yn, l = e.f;
	Un = null, Wn = 0, Gn = null, V = l & 96 ? null : e, Vn = null, Ue(e.ctx), Rn = !1, Yn = ++Jn, e.ac !== null && (it(() => {
		e.ac.abort(ge);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = tr(e);
		if (qe() && Gn !== null && !Rn && f !== null && !(e.f & 6146)) for (var p = 0; p < Gn.length; p++) $n(Gn[p], e);
		if (i !== null && i !== e) {
			if (Jn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Jn;
			if (t !== null) for (let e of t) e.rv = Jn;
			Gn !== null && (r === null ? r = Gn : r.push(...Gn));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return tr(e), ln(t);
	} finally {
		e.f ^= ae, Un = t, Wn = n, Gn = r, V = i, Vn = a, Ue(o), Rn = s, Yn = c;
	}
}
function tr(e) {
	var t = e.deps, n = P?.is_fork;
	if (Un !== null) {
		var r;
		if (n || rr(e, Wn), t !== null && Wn > 0) for (t.length = Wn + Un.length, r = 0; r < Un.length; r++) t[Wn + r] = Un[r];
		else e.deps = t = Un;
		if (mn() && e.f & 512) for (r = Wn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Wn < t.length && (rr(e, Wn), t.length = Wn);
	return t;
}
function nr(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = o.call(r, e);
		if (i !== -1) {
			var a = r.length - 1;
			a === 0 ? r = n.reactions = null : (r[i] = r[a], r.pop());
		}
	}
	if (r === null && n.f & 2 && (Un === null || !s.call(Un, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && $e(c), c.ac !== null && it(() => {
			c.ac.abort(ge), c.ac = null, M(c, x);
		}), _t(c), rr(c, 0);
	}
}
function rr(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) nr(e, n[r]);
}
function ir(e) {
	var t = e.f;
	if (!(t & 16384)) {
		M(e, b);
		var n = H;
		H = e;
		try {
			t & 16777232 ? En(e) : Tn(e), wn(e);
			var r = er(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = qn;
		} finally {
			H = n;
		}
	}
}
async function ar() {
	await Promise.resolve(), At();
}
function U(e) {
	var t = !!(e.f & 2);
	if (Fn?.add(e), V !== null && !Rn && !(H !== null && H.f & 16384) && (Vn === null || !Vn.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < Jn && (e.rv = Jn, Un === null && n !== null && n[Wn] === e ? Wn++ : Un === null ? Un = [e] : Un.push(e));
		else {
			V.deps ??= [], s.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : s.call(r, V) || r.push(V);
		}
	}
	if (In && Rt.has(e)) return Rt.get(e);
	if (t) {
		var i = e;
		if (In) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || sr(i)) && (a = ht(i)), Rt.set(i, a), a;
		}
		var o = !(i.f & 512) && !Rn && V !== null && !!(V.f & 512), c = (i.f & ee) === 0;
		Qn(i) && (o && (i.f |= 512), gt(i)), o && !c && (vt(i), or(i));
	}
	if (xt?.has(e)) return xt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function or(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : s.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (vt(n), or(n));
	}
}
function sr(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Rt.has(t) || t.f & 2 && sr(t)) return !0;
	return !1;
}
function cr(e) {
	var t = Rn;
	try {
		return Rn = !0, e();
	} finally {
		Rn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var lr = Symbol("events"), ur = /* @__PURE__ */ new Set(), dr = /* @__PURE__ */ new Set();
function fr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || vr.call(t, e), !e.cancelBubble) return it(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, Xe(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function pr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = fr(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && hn(() => {
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
		var c = i.indexOf(s);
		if (c !== -1 && (t === document || t === window)) {
			e[lr] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		c <= u && (o = c);
	}
	if (a = i[o] || e.target, a !== t) {
		l(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = V, f = H;
		zn(null), Bn(null);
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
			e[lr] = t, delete e.currentTarget, zn(d), Bn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var yr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function br(e) {
	return yr?.createHTML(e) ?? e;
}
function xr(e) {
	var t = sn("template");
	return t.innerHTML = br(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Sr(e, t) {
	var n = H;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function W(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (D) return Sr(O, null), O;
		i === void 0 && (i = xr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ tn(i)));
		var t = r || Xt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ tn(t), s = t.lastChild;
			Sr(o, s);
		} else Sr(t, t);
		return t;
	};
}
function Cr(e = "") {
	if (!D) {
		var t = en(e + "");
		return Sr(t, t), t;
	}
	var n = O;
	return n.nodeType === 3 ? cn(n) : (n.before(n = en()), Se(n)), Sr(n, n), n;
}
function wr() {
	if (D) return Sr(O, null), O;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = en();
	return e.append(t, n), Sr(t, n), e;
}
function G(e, t) {
	if (D) {
		var n = H;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), k();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var Tr = ["touchstart", "touchmove"];
function Er(e) {
	return Tr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Dr(e) {
	let t = 0, n = Bt(0), r;
	return () => {
		mn() && (U(n), xn(() => (t === 0 && (r = cr(() => e(() => Kt(n)))), t += 1, () => {
			Xe(() => {
				--t, t === 0 && (r?.(), r = void 0, Kt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Or = T | ne;
function kr(e, t, n, r) {
	new Ar(e, t, n, r);
}
var Ar = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = D ? O : null;
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
	#m = Dr(() => (this.#p = Bt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = H;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = H.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = Sn(() => {
			if (D) {
				let e = we(this.#t);
				k();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Or), D && (this.#e = O);
	}
	#h() {
		try {
			this.#a = Cn(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Xe(r), t && (this.#s = Cn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? be() : (t = !0, n && Re(), this.#s !== null && An(this.#s, () => {
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
					un(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (w | te)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = Cn(() => e(this.#e)), Xe(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = en(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return Cn(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							un(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(P)) : this.#u === 0 && (this.#e.before(e), this.#c = null, An(this.#o, () => {
					this.#o = null;
				}), this.#x(P));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = Cn(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Pn(this.#a, e);
				let t = this.#n.pending;
				this.#o = Cn(() => t(this.#e));
			} else this.#x(P);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		et(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = H, n = V, r = He;
		Bn(this.#i), zn(this.#i), Ue(this.#i.ctx);
		try {
			return kt.ensure(), e();
		} finally {
			Bn(t), zn(n), Ue(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && An(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Xe(() => {
			this.#d = !1, this.#p && Wt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), U(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		P?.is_fork ? (this.#a && P.skip_effect(this.#a), this.#o && P.skip_effect(this.#o), this.#s && P.skip_effect(this.#s), P.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (Dn(this.#a), null), this.#o &&= (Dn(this.#o), null), this.#s &&= (Dn(this.#s), null), D && (Se(this.#t), j(), Se(Ce()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return Cn(() => {
						var r = H;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return un(e, this.#i.parent), null;
				}
			}));
		};
		Xe(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					un(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => un(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function K(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[he] ??= e.nodeValue) && (e[he] = n, e.nodeValue = `${n}`);
}
function jr(e, t) {
	return Nr(e, t);
}
var Mr = /* @__PURE__ */ new Map();
function Nr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: s = !0, transformError: l }) {
	$t();
	var u = void 0, d = vn(() => {
		var s = r ?? n.appendChild(en());
		kr(s, { pending: () => {} }, (n) => {
			We({});
			var r = He;
			if (o && (r.c = o), a && (i.$$events = a), D && Sr(n, null), u = t(n, i) || Ke(), D && (H.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw ye(), e;
			Ge();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Er(r);
					for (let e of [n, document]) {
						var a = Mr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Mr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, vr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(c(ur)), dr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Mr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, vr), t.delete(e), t.size === 0 && Mr.delete(r)) : t.set(e, i);
			}
			dr.delete(f), s !== r && s.parentNode?.removeChild(s);
		};
	});
	return Pr.set(u, d), u;
}
var Pr = /* @__PURE__ */ new WeakMap();
function Fr(e, t) {
	let n = Pr.get(e);
	return n ? (Pr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Ir = class {
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
			if (n) Mn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Mn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (Dn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Pn(r, t), t.append(en()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else Dn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), An(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (Dn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = P, r = on();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = en();
				i.append(a), this.#n.set(e, {
					effect: Cn(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, Cn(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else D && (this.anchor = O), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function q(e, t, n = !1) {
	var r;
	D && (r = O, k());
	var i = new Ir(e), a = n ? T : 0;
	function o(e, t) {
		if (D) {
			var n = we(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Ce();
				Se(a), i.anchor = a, xe(!1), i.ensure(e, t), xe(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	Sn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function Lr(e, t) {
	return t;
}
function Rr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, s = 0; s < i; s++) {
		let n = t[s];
		An(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					zr(e, c(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --o;
		}, !1);
	}
	if (o === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			an(d), d.append(u), e.items.clear();
		}
		zr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function zr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ie, Pn(a, document.createDocumentFragment())) : Dn(t[i], n);
	}
}
var Br;
function Vr(e, t, n, r, i, o = null) {
	var s = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		s = D ? Se(/* @__PURE__ */ tn(u)) : u.appendChild(en());
	}
	D && k();
	var d = null, f = /* @__PURE__ */ N(() => {
		var e = n();
		return a(e) ? e : e == null ? [] : c(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = U(f);
			_.fallback = d, Ur(_, n, s, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, Gr(d, null, s)) : Mn(d) : An(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: Sn(() => {
			var e = U(f), a = e.length;
			let c = !1;
			D && we(s) === "[!" != (a === 0) && (s = Ce(), Se(s), xe(!1), c = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = P, v = on(), y = 0; y < a; y += 1) {
				D && O.nodeType === 8 && O.data === "]" && (s = O, c = !0, xe(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && Wt(S.v, b), S.i && Wt(S.i, y), v && _.unskip_effect(S.e)) : (S = Wr(l, m ? s : Br ??= en(), b, x, y, i, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (a === 0 && o && !d && (m ? d = Cn(() => o(s)) : (d = Cn(() => o(Br ??= en())), d.f |= ie)), a > u.size && ke("", "", ""), D && a > 0 && Se(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			c && xe(!0), U(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, D && (s = O);
}
function Hr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Ur(e, t, n, r, i) {
	var a = !!(r & 8), o = t.length, s = e.items, l = Hr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < o; v += 1) h = t[v], g = i(h, v), _ = s.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < o; v += 1) {
		if (h = t[v], g = i(h, v), _ = s.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (Mn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ie, _ === l) Gr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Kr(e, d, _), Kr(e, _, y), Gr(_, y, n), d = _, p = [], m = [], l = Hr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Gr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Kr(e, S.prev, C.next), Kr(e, d, S), Kr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Gr(_, l, n), Kr(e, _.prev, _.next), Kr(e, _, d === null ? e.effect.first : d.next), Kr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Hr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Hr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (zr(e, c(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = Hr(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && o === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			Rr(e, w, te);
		}
	}
	a && Xe(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Wr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Bt(n) : /* @__PURE__ */ Vt(n, !1, !1) : null, l = o & 2 ? Bt(i) : null;
	return {
		v: c,
		i: l,
		e: Cn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Gr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ nn(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Kr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
function qr(t, n, a = !1, o = !1, s = !1, c = !1) {
	var l = t, u = "";
	if (a) {
		var d = t;
		D && (l = Se(/* @__PURE__ */ tn(d)));
	}
	B(() => {
		var t = H;
		if (u === (u = n() ?? "")) D && k();
		else if (a && !D) t.nodes = null, d.innerHTML = u, u !== "" && Sr(/* @__PURE__ */ tn(d), d.lastChild);
		else if (t.nodes !== null && (On(t.nodes.start, t.nodes.end), t.nodes = null), u !== "") {
			if (D) {
				for (var c = O.data, f = k(), p = f; f !== null && (f.nodeType !== 8 || f.data !== "");) p = f, f = /* @__PURE__ */ nn(f);
				if (f === null) throw ye(), e;
				Sr(O, p), l = Se(f);
			} else {
				var m = sn(o ? "svg" : s ? "math" : "template", o ? r : s ? i : void 0);
				m.innerHTML = u;
				var h = o || s ? m : m.content;
				if (Sr(/* @__PURE__ */ tn(h), h.lastChild), o || s) for (; /* @__PURE__ */ tn(h);) l.before(/* @__PURE__ */ tn(h));
				else l.before(h);
			}
		}
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var Jr = [..." 	\n\r\f\xA0\v﻿"];
function Yr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Jr.includes(r[o - 1])) && (s === r.length || Jr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Xr(e, t, n, r, i, a) {
	var o = e[pe];
	if (D || o !== n || o === void 0) {
		var s = Yr(n, r, a);
		D && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[pe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Zr = Symbol("is custom element"), Qr = Symbol("is html"), $r = _e ? "link" : "LINK";
function ei(e) {
	if (D) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					ti(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					ti(e, "checked", null), e.checked = r;
				}
			}
		};
		e[E] = n, Xe(n), rt();
	}
}
function ti(e, t, n, r) {
	var i = ni(e);
	if (!(D && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === $r))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[de] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && ii(e).has(t) ? e[t] = n : (!D || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function ni(e) {
	return e[fe] ??= {
		[Zr]: e.nodeName.includes("-"),
		[Qr]: e.namespaceURI === n
	};
}
var ri = /* @__PURE__ */ new Map();
function ii(e) {
	var t = e.getAttribute("is") || e.nodeName, n = ri.get(t);
	if (n) return n;
	ri.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = d(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = m(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/input.js
function ai(e, t, n = t) {
	var r = /* @__PURE__ */ new WeakSet();
	at(e, "input", async (i) => {
		var a = i ? e.defaultValue : e.value;
		if (a = si(e) ? ci(a) : a, n(a), P !== null && r.add(P), await ar(), a !== (a = t())) {
			var o = e.selectionStart, s = e.selectionEnd, c = e.value.length;
			if (e.value = a ?? "", s !== null) {
				var l = e.value.length;
				o === s && s === c && l > c ? (e.selectionStart = l, e.selectionEnd = l) : (e.selectionStart = o, e.selectionEnd = Math.min(s, l));
			}
		}
	}), (D && e.defaultValue !== e.value || cr(t) == null && e.value) && (n(si(e) ? ci(e.value) : e.value), P !== null && r.add(P)), xn(() => {
		var n = t();
		if (e === document.activeElement) {
			var i = P;
			if (r.has(i)) return;
		}
		si(e) && n === ci(e.value) || (e.type !== "date" || n || e.value) && n !== e.value && (e.value = n ?? "");
	});
}
function oi(e, t, n = t) {
	at(e, "change", (t) => {
		n(t ? e.defaultChecked : e.checked);
	}), (D && e.defaultChecked !== e.checked || cr(t) == null) && n(e.checked), xn(() => {
		e.checked = !!t();
	});
}
function si(e) {
	var t = e.type;
	return t === "number" || t === "range";
}
function ci(e) {
	return e === "" ? null : +e;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/this.js
function li(e, t) {
	return e === t || e?.[ce] === t;
}
function ui(e = Ke(), t, n, r) {
	var i = He.r, a = H;
	return yn(() => {
		var o, s;
		return xn(() => {
			o = s, s = r?.() || [], cr(() => {
				li(n(...s), e) || (t(e, ...s), o && li(n(...o), e) && t(null, ...o));
			});
		}), () => {
			let r = a;
			for (; r !== i && r.parent !== null && r.parent.f & 33554432;) r = r.parent;
			let o = () => {
				s && li(n(...s), e) && t(null, ...s);
			}, c = r.teardown;
			r.teardown = () => {
				o(), c?.();
			};
		};
	}), e;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var di = !1;
function fi(e) {
	var t = di;
	try {
		return di = !1, [e(), di];
	} finally {
		di = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function pi(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, c = !0, l = void 0, d = () => o && i ? (l ??= /* @__PURE__ */ ut(r), U(l)) : (c && (c = !1, s = o ? cr(r) : r), s);
	let f;
	if (a) {
		var p = ce in e || ue in e;
		f = u(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = fi(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Pe(t), f(m)));
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
	var v = !1, y = (n & 1 ? ut : N)(() => (v = !1, g()));
	a && U(y);
	var b = H;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? U(y) : i && a ? Jt(e) : e;
			return I(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return In && v || b.f & 16384 ? y.v : U(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// plugins/chat/transcript.svelte.js
var mi = "Messages for you:";
function hi(e) {
	switch (e) {
		case "pending": return "input";
		case "in_progress": return "running";
		case "completed": return "done";
		case "failed": return "failed";
		default: return e || "input";
	}
}
function gi(e) {
	if (!e.startsWith(mi)) return {
		source: "controller",
		text: e
	};
	let t = /\n\nFrom ([^\s(]+) \(/.exec(e), n = t ? t[1] : "", r = e.slice(17).replace(/^\s*\n\nFrom [^\n]*\n/, "").replace(/\n\nFrom [^\n]*\n/g, "\n\n");
	return {
		source: n.startsWith("desk:") ? "desk" : n ? "operator" : "message",
		text: r.trim(),
		who: n
	};
}
var _i = class {
	#e = /* @__PURE__ */ F(Jt([]));
	get blocks() {
		return U(this.#e);
	}
	set blocks(e) {
		I(this.#e, e, !0);
	}
	#t = /* @__PURE__ */ F(!1);
	get busy() {
		return U(this.#t);
	}
	set busy(e) {
		I(this.#t, e, !0);
	}
	#n = /* @__PURE__ */ F(null);
	get usage() {
		return U(this.#n);
	}
	set usage(e) {
		I(this.#n, e, !0);
	}
	last(e) {
		let t = this.blocks[this.blocks.length - 1];
		return t && t.kind === e ? t : void 0;
	}
	tool(e) {
		for (let t = this.blocks.length - 1; t >= 0; t--) {
			let n = this.blocks[t];
			if (n.kind === "tool" && n.id === e) return n;
		}
	}
	settle() {
		for (let e of this.blocks) (e.kind === "assistant" || e.kind === "reasoning") && e.live && (e.live = !1);
	}
	apply(e) {
		let t = e.seq || 0;
		switch (e.type) {
			case "turn_start": {
				this.settle(), this.busy = !0;
				let n = gi(e.text || "");
				e.context && this.blocks.push({
					kind: "system",
					text: "Context given to the session:\n" + e.context,
					seq: t
				}), n.text && this.blocks.push({
					kind: "user",
					text: n.text,
					source: n.source,
					context: e.context || "",
					seq: t
				});
				break;
			}
			case "text": {
				this.busy = !0;
				let n = this.last("assistant");
				n?.live ? n.text += e.text : this.blocks.push({
					kind: "assistant",
					text: e.text || "",
					live: !0,
					seq: t
				});
				break;
			}
			case "reasoning": {
				this.busy = !0;
				let n = this.last("reasoning");
				n?.live ? n.text += e.text : this.blocks.push({
					kind: "reasoning",
					text: e.text || "",
					live: !0,
					seq: t
				});
				break;
			}
			case "tool": {
				let n = e.tool;
				if (!n) break;
				this.busy = !0;
				let r = n.id || n.title, i = this.tool(r);
				i ? (n.title && (i.name = n.title), i.state = hi(n.status), n.input && (i.input = n.input), n.output && (i.output = n.output)) : (this.settle(), this.blocks.push({
					kind: "tool",
					id: r,
					name: n.title || r,
					input: n.input || "",
					output: n.output || "",
					state: hi(n.status),
					seq: t
				}));
				break;
			}
			case "turn_end":
				this.settle(), this.busy = !1, e.error ? this.blocks.push({
					kind: "error",
					text: "Turn failed: " + e.error,
					seq: t
				}) : e.stop_reason === "cancelled" ? this.blocks.push({
					kind: "system",
					text: "Stopped",
					seq: t
				}) : e.stop_reason && e.stop_reason !== "end_turn" && this.blocks.push({
					kind: "system",
					text: "Turn ended: " + e.stop_reason,
					seq: t
				});
				break;
			case "usage":
				e.usage && (this.usage = e.usage);
				break;
			case "state":
				e.status === "closed" && (this.busy = !1, this.blocks.push({
					kind: "system",
					text: "Session closed" + (e.resolution ? " · " + e.resolution : ""),
					seq: t
				}));
				break;
			default: return !1;
		}
		return !0;
	}
}, vi = class {
	#e = /* @__PURE__ */ F("connecting");
	get status() {
		return U(this.#e);
	}
	set status(e) {
		I(this.#e, e, !0);
	}
	#t = /* @__PURE__ */ F(0);
	get first() {
		return U(this.#t);
	}
	set first(e) {
		I(this.#t, e, !0);
	}
	#n = /* @__PURE__ */ F(!1);
	get loadingOlder() {
		return U(this.#n);
	}
	set loadingOlder(e) {
		I(this.#n, e, !0);
	}
	#r = /* @__PURE__ */ F("");
	get error() {
		return U(this.#r);
	}
	set error(e) {
		I(this.#r, e, !0);
	}
	#i = /* @__PURE__ */ F(0);
	get version() {
		return U(this.#i);
	}
	set version(e) {
		I(this.#i, e, !0);
	}
	#a = new _i();
	#o = null;
	get blocks() {
		return this.#a.blocks;
	}
	get busy() {
		return this.#a.busy;
	}
	set busy(e) {
		this.#a.busy = e;
	}
	get usage() {
		return this.#a.usage;
	}
	constructor(e, t) {
		this.api = e, this.id = t;
	}
	open(e = 200) {
		this.close(), this.#o = this.api.session.stream(this.id, { tail: e }, (e) => {
			e.type === "open" ? this.status = "live" : e.type === "end" ? (this.status = "closed", this.busy = !1) : e.type === "error" ? this.status !== "closed" && (this.status = "reconnecting") : e.type === "event" && this.#s(e.event);
		});
	}
	close() {
		this.#o?.(), this.#o = null;
	}
	#s(e) {
		e.seq && (!this.first || e.seq < this.first) && (this.first = e.seq), this.#a.apply(e) && this.version++;
	}
	get earlier() {
		return this.first > 1 ? this.first - 1 : 0;
	}
	async loadOlder() {
		if (!this.loadingOlder && this.earlier) {
			this.loadingOlder = !0, this.error = "";
			try {
				let e = await this.api.session.older(this.id, this.first, 400), t = new _i();
				for (let n of e.events || []) t.apply(n);
				t.settle(), this.#a.blocks = [...Be(t.blocks), ...this.#a.blocks], this.first = Math.max(1, e.first || 1), this.version++;
			} catch (e) {
				this.error = String(e.message || e);
			} finally {
				this.loadingOlder = !1;
			}
		}
	}
	prompts() {
		return this.blocks.filter((e) => e.kind === "user" && e.source !== "controller").map((e) => e.text);
	}
};
// node_modules/marked/lib/marked.esm.js
function yi() {
	return {
		async: !1,
		breaks: !1,
		extensions: null,
		gfm: !0,
		hooks: null,
		pedantic: !1,
		renderer: null,
		silent: !1,
		tokenizer: null,
		walkTokens: null
	};
}
var bi = yi();
function xi(e) {
	bi = e;
}
var Si = { exec: () => null };
function Ci(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function J(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(Y.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var wi = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), Y = {
	codeRemoveIndent: /^(?: {0,3}\t| {1,4})/gm,
	outputLinkReplace: /\\([\[\]])/g,
	indentCodeCompensation: /^(\s+)(?:```)/,
	beginningSpace: /^\s+/,
	endingHash: /#$/,
	startingSpaceChar: /^ /,
	endingSpaceChar: / $/,
	endingSpaceTabChar: /[ \t]$/,
	nonSpaceChar: /[^ ]/,
	newLineCharGlobal: /\n/g,
	tabCharGlobal: /\t/g,
	leadingSpaceTab: /^[ \t]+/,
	multipleSpaceGlobal: /\s+/g,
	blankLine: /^[ \t]*$/,
	doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
	blockquoteStart: /^ {0,3}>/,
	blockquoteSetextReplace: /\n {0,3}((?:=+|-+)[ \t]*)(?=\n|$)/g,
	blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
	blockquoteMarkerWhitespace: /^( {0,3}>)([ \t]+)/gm,
	listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
	listIsTask: /^\[[ xX]\][ \t]+\S/,
	listReplaceTask: /^\[[ xX]\][ \t]+/,
	listTaskCheckbox: /\[[ xX]\]/,
	anyLine: /\n.*\n/,
	hrefBrackets: /^<(.*)>$/,
	tableDelimiter: /[:|]/,
	tableAlignChars: /^\||\| *$/g,
	tableRowBlankLine: /\n[ \t]*$/,
	tableAlignRight: /^ *-+: *$/,
	tableAlignCenter: /^ *:-+: *$/,
	tableAlignLeft: /^ *:-+ *$/,
	startATag: /^<a /i,
	endATag: /^<\/a>/i,
	startPreScriptTag: /^<(pre|code|kbd|script)(\s|>)/i,
	endPreScriptTag: /^<\/(pre|code|kbd|script)(\s|>)/i,
	startAngleBracket: /^</,
	endAngleBracket: />$/,
	pedanticHrefTitle: /^([^'"]*[^\s])\s+(['"])(.*)\2/,
	unicodeAlphaNumeric: /[\p{L}\p{N}]/u,
	numericCharacterReference: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));/,
	numericCharacterReferenceOrAmpersand: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	escapeTest: /[&<>"']/,
	escapeReplace: /[&<>"']/g,
	escapeTestNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/,
	escapeReplaceNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	caret: /(^|[^\[])\^/g,
	percentDecode: /%25/g,
	findPipe: /\|/g,
	splitPipe: / \|/,
	slashPipe: /\\\|/g,
	carriageReturn: /\r\n|\r/g,
	spaceLine: /^ +$/gm,
	notSpaceStart: /^\S*/,
	endingNewline: /\n$/,
	listItemRegex: (e) => RegExp(`^( {0,3}${e})((?:[	 ][^\\n]*)?(?:\\n|$))`),
	nextBulletRegex: Ci((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Ci((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Ci((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Ci((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Ci((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Ci((e) => RegExp(`^ {0,${e}}(?:</?(?:${zi})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Ci((e) => RegExp(`^ {0,${e}}>`))
}, Ti = /^(?:[ \t]*(?:\n|$))+/, Ei = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Di = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, Oi = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, ki = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, Ai = / {0,3}(?:[*+-]|\d{1,9}[.)])/, ji = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, Mi = J(ji).replace(/bull/g, Ai).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Ni = J(ji).replace(/bull/g, Ai).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), Pi = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, Fi = /^[^\n]+/, Ii = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Li = J(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", Ii).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), Ri = J(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, Ai).getRegex(), zi = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Bi = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Vi = J("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Bi).replace("tag", zi).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Hi = (e) => J(Pi).replace("hr", Oi).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", zi).getRegex(), Ui = Hi(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), Wi = Hi(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Gi = {
	blockquote: J(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", Wi).getRegex(),
	code: Ei,
	def: Li,
	fences: Di,
	heading: ki,
	hr: Oi,
	html: Vi,
	lheading: Mi,
	list: Ri,
	newline: Ti,
	paragraph: Ui,
	table: Si,
	text: Fi
}, Ki = J("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", Oi).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", zi).getRegex(), qi = {
	...Gi,
	lheading: Ni,
	table: Ki,
	paragraph: J(Pi).replace("hr", Oi).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Ki).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", zi).getRegex()
}, Ji = {
	...Gi,
	html: J("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Bi).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: Si,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: J(Pi).replace("hr", Oi).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Mi).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Yi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Xi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Zi = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Qi = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, $i = /[\p{P}\p{S}]/u, ea = /[\s\p{P}\p{S}]/u, ta = /[^\s\p{P}\p{S}]/u, na = J(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, ea).getRegex(), ra = /[\p{Pi}\p{Ps}"']/u, ia = /(?!~)[\p{P}\p{S}]/u, aa = /(?!~)[\s\p{P}\p{S}]/u, oa = /(?:[^\s\p{P}\p{S}]|~)/u, sa = J(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", wi ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), ca = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, la = J(ca, "u").replace(/punct/g, $i).getRegex(), ua = J(ca, "u").replace(/punct/g, ia).getRegex(), da = J(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, ra).replace(/punct/g, $i).getRegex(), fa = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", pa = J(fa, "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ma = J(fa, "gu").replace(/notPunctSpace/g, oa).replace(/punctSpace/g, aa).replace(/punct/g, ia).getRegex(), ha = J("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ga = J("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), _a = J("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), va = J(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, $i).getRegex(), ya = J("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ba = J(/\\(punct)/, "gu").replace(/punct/g, $i).getRegex(), xa = J(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, $i).getRegex(), Sa = J(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), Ca = J(Bi).replace("(?:-->|$)", "-->").getRegex(), wa = J("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", Ca).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Ta = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Ea = J(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Ta).getRegex(), Da = J(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Ea).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), Oa = J(/^!?\[(label)\]\[(ref)\]/).replace("label", Ea).replace("ref", Ii).getRegex(), ka = J(/^!?\[(ref)\](?:\[\])?/).replace("ref", Ii).getRegex(), Aa = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, ja = J(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Ta).getRegex(), Ma = J("reflink|nolink(?!\\()", "g").replace("reflink", J(/^!?\[(label)\]\[(ref)\]/).replace("label", ja).replace("ref", Aa).getRegex()).replace("nolink", J(/^!?\[(ref)\](?:\[\])?/).replace("ref", Aa).getRegex()).getRegex(), Na = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, Pa = J(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), Fa = {
	_backpedal: Si,
	anyPunctuation: ba,
	anyPunctuationOutsideCode: xa,
	autolink: Sa,
	blockSkip: sa,
	br: Zi,
	code: Xi,
	del: Si,
	delLDelim: Si,
	delRDelim: Si,
	emStrongLDelim: la,
	emStrongRDelimAst: pa,
	emStrongRDelimUnd: ga,
	escape: Yi,
	link: Da,
	nolink: ka,
	punctuation: na,
	reflink: Oa,
	reflinkSearch: Ma,
	tag: wa,
	text: Qi,
	url: Si
}, Ia = {
	...Fa,
	emStrongLDelim: da,
	emStrongRDelimAst: ha,
	emStrongRDelimUnd: _a,
	link: J(/^!?\[(label)\]\((.*?)\)/).replace("label", Ea).getRegex(),
	reflink: J(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Ea).getRegex()
}, La = {
	...Fa,
	emStrongRDelimAst: ma,
	emStrongLDelim: ua,
	delLDelim: va,
	delRDelim: ya,
	url: J(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", Pa).replace("protocol", Na).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: J(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", Na).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", wi ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, Ra = {
	...La,
	br: J(Zi).replace("{2,}", "*").getRegex(),
	text: J(La.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, za = {
	normal: Gi,
	gfm: qi,
	pedantic: Ji
}, Ba = {
	normal: Fa,
	gfm: La,
	breaks: Ra,
	pedantic: Ia
}, Va = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, Ha = (e) => Va[e];
function Ua(e, t) {
	if (t) {
		if (Y.escapeTest.test(e)) return e.replace(Y.escapeReplace, Ha);
	} else if (Y.escapeTestNoEncode.test(e)) return e.replace(Y.escapeReplaceNoEncode, Ha);
	return e;
}
function Wa(e) {
	return Y.numericCharacterReference.test(e) ? e.replace(Y.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Ga(e) {
	try {
		e = encodeURI(e).replace(Y.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Ka(e, t) {
	let n = e.replace(Y.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(Y.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(Y.slashPipe, "|");
	return n;
}
function qa(e, t, n) {
	let r = e.length;
	if (r === 0) return "";
	let i = 0;
	for (; i < r;) {
		let a = e.charAt(r - i - 1);
		if (a === t && !n) i++;
		else if (a !== t && n) i++;
		else break;
	}
	return e.slice(0, r - i);
}
function Ja(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && Y.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Ya(e) {
	return e.trim().replace(Y.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Xa(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Za = /* @__PURE__ */ new WeakMap();
function Qa(e) {
	let t = {
		"*": -1,
		_: -1
	}, n = -1, r = e.length;
	for (let i = 0; i < r; i++) {
		let a = e[i];
		if (a !== "*" && a !== "_" && a !== "~") continue;
		let o = i === 0 ? "\n" : e[i - 1], s = i + 1;
		for (; s < r && e[s] === a;) s++;
		o !== " " && o !== "	" && o !== "\n" && o !== "\r" && o !== "\f" && (a === "~" ? n = i : t[a] = i), i = s - 1;
	}
	return {
		maskedSrc: e,
		lastPossibleEmStrongCloserIndex: t,
		lastPossibleDelCloserIndex: n
	};
}
function $a(e, t) {
	let n = Qa(t), r = Za.get(e);
	r ? r.push(n) : Za.set(e, [n]);
}
function eo(e) {
	let t = Za.get(e);
	t && (t.pop(), t.length === 0 && Za.delete(e));
}
function to(e, t, n) {
	let r = Za.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : Qa(t).lastPossibleEmStrongCloserIndex[n];
}
function no(e, t) {
	let n = Za.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : Qa(t).lastPossibleDelCloserIndex;
}
function ro(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function io(e, t, n, r, i) {
	let a = t.href, o = t.title || null, s = e[1].replace(i.other.outputLinkReplace, "$1"), c = e[0].charAt(0) === "!", l = r.state.inLink, u = r.state.linkEmitted, d = r.state.inRawBlock;
	r.state.inLink = !0, r.state.linkEmitted = !1;
	let f = r.inlineTokens(s), p = r.state.linkEmitted;
	if (r.state.linkEmitted = u, r.state.inLink = l, !c) {
		if (p) {
			r.state.inRawBlock = d;
			return;
		}
		r.state.linkEmitted = !0;
	}
	return {
		type: c ? "image" : "link",
		raw: n,
		href: a,
		title: o,
		text: s,
		tokens: f
	};
}
function ao(e, t, n) {
	let r = e.match(n.other.indentCodeCompensation);
	if (r === null) return t;
	let i = r[1];
	return t.split("\n").map((e) => {
		let t = e.match(n.other.beginningSpace);
		if (t === null) return e;
		let [r] = t;
		return e.slice(Math.min(r.length, i.length));
	}).join("\n");
}
function oo(e, t, n, r) {
	if (!t.includes("<")) return !1;
	for (let i = 0; i < t.length; i++) {
		if (t[i] === "\\") {
			i++;
			continue;
		}
		if (t[i] === "`") {
			let e = r.inline.code.exec(t.slice(i));
			if (e) {
				i += e[0].length - 1;
				continue;
			}
		}
		if (t[i] !== "<") continue;
		let a = e.slice(n + i), o = r.inline.tag.exec(a) || r.inline.autolink.exec(a);
		if (o) {
			if (o[0].length > t.length - i) return !0;
			i += o[0].length - 1;
		}
	}
	return !1;
}
var so = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || bi;
	}
	space(e) {
		let t = this.rules.block.newline.exec(e);
		if (t && t[0].length > 0) return {
			type: "space",
			raw: t[0]
		};
	}
	code(e) {
		let t = this.rules.block.code.exec(e);
		if (t) {
			let e = this.options.pedantic ? t[0] : Ja(t[0]);
			return {
				type: "code",
				raw: e,
				codeBlockStyle: "indented",
				text: e.replace(this.rules.other.codeRemoveIndent, "")
			};
		}
	}
	fences(e) {
		let t = this.rules.block.fences.exec(e);
		if (t) {
			let e = t[0], n = ao(e, t[3] || "", this.rules);
			return {
				type: "code",
				raw: e,
				lang: t[2] ? t[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : t[2],
				text: n
			};
		}
	}
	heading(e) {
		let t = this.rules.block.heading.exec(e);
		if (t) {
			let e = t[2].trim();
			if (this.rules.other.endingHash.test(e)) {
				let t = qa(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: qa(t[0], "\n"),
				depth: t[1].length,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	hr(e) {
		let t = this.rules.block.hr.exec(e);
		if (t) return {
			type: "hr",
			raw: qa(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = qa(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + ro(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
				n = n ? `${n}
${s}` : s, r = r ? `${r}
${c}` : c;
				let l = this.lexer.state.top;
				if (this.lexer.state.top = !0, this.lexer.blockTokens(c, i, !0), this.lexer.state.top = l, e.length === 0) break;
				let u = i.at(-1);
				if (u?.type === "code" || u?.type === "heading" || u?.type === "hr" || u?.type === "html" || u?.type === "table") break;
				if (u?.type === "blockquote") {
					let t = u, a = e.join("\n"), o = t.raw + "\n" + a.replace(this.rules.other.blockquoteSetextReplace2, ""), s = this.blockquote(o);
					i[i.length - 1] = s;
					let c = o.substring(s.raw.length).replace(/^\n/, ""), l = c ? c.split("\n").length : 0, d = l ? e.slice(0, -l) : e;
					if (d.length > 0 && (n = `${n}
${d.join("\n")}`), r = r.substring(0, r.length - t.text.length) + s.text, e = e.slice(d.length), e.length > 0 && this.rules.other.blockquoteStart.test(e[0])) continue;
					break;
				}
				if (u?.type === "list") {
					let t = u, a = t.raw + "\n" + e.join("\n"), o = this.list(a);
					i[i.length - 1] = o;
					let s = o.raw.length > t.raw.length ? a.slice(t.raw.length + 1, o.raw.length).split("\n").length : 0;
					if (s > 0) {
						let t = e.slice(0, s).join("\n");
						n = `${n}
${t}`, r = `${r}
${t}`;
					}
					e = e.slice(s);
				}
			}
			return {
				type: "blockquote",
				raw: n,
				tokens: i,
				text: r
			};
		}
	}
	list(e) {
		let t = this.rules.block.list.exec(e);
		if (t) {
			let n = e, r = () => n.length - e.length, i = t[1].trim(), a = i.length > 1, o = {
				type: "list",
				raw: "",
				ordered: a,
				start: a ? +i.slice(0, -1) : "",
				loose: !1,
				items: []
			};
			i = a ? `\\d{1,9}\\${i.slice(-1)}` : `\\${i}`, this.options.pedantic && (i = a ? i : "[*+-]");
			let s = this.rules.other.listItemRegex(i), c = !1;
			for (; e;) {
				let i = !1, a = "";
				if (!(t = s.exec(e)) || this.rules.block.hr.test(e)) break;
				let l = r();
				e = e.substring(t[0].length);
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? ro(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => ro(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => ro(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
						if (l.search(this.rules.other.nonSpaceChar) >= h || !p.trim()) a += "\n" + l.slice(h);
						else {
							if (m || f.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4 || r.test(f) || i.test(f) || o.test(f) || n.test(f)) break;
							let e = !this.options.pedantic && /^ {0,3}(?:=+|-+) *$/.test(p) && this.rules.block.lheading.test(`${f}
${p}`) ? `    ${l.replace(/^ {0,3}/, "")}` : p;
							a += "\n" + e;
						}
						m = !p.trim(), e = e.substring(c.length + 1), f = l.slice(h);
					}
				}
				let g = n.slice(l, r());
				o.loose || (c ? o.loose = !0 : this.rules.other.doubleBlankLine.test(g) && (c = !0)), o.items.push({
					type: "list_item",
					raw: g,
					task: !!this.options.gfm && this.rules.other.listIsTask.test(a),
					loose: !1,
					text: a,
					tokens: []
				});
			}
			let l = o.items.at(-1);
			if (l) l.raw = l.raw.trimEnd(), l.text = l.text.trimEnd();
			else return;
			o.raw = n.slice(0, r()).trimEnd();
			for (let e of o.items) if (this.lexer.state.top = !1, e.tokens = this.lexer.blockTokens(e.text, []), !o.loose) {
				let t = e.tokens.filter((e) => e.type === "space");
				o.loose = t.length > 0 && t.some((e) => this.rules.other.anyLine.test(e.raw));
			}
			for (let e of o.items) {
				let t = e.tokens[0];
				if (e.task && (t?.type === "text" || t?.type === "paragraph")) {
					e.text = e.text.replace(this.rules.other.listReplaceTask, ""), t.raw = t.raw.replace(this.rules.other.listReplaceTask, ""), t.text = t.text.replace(this.rules.other.listReplaceTask, "");
					for (let e = this.lexer.inlineQueue.length - 1; e >= 0; e--) if (this.rules.other.listIsTask.test(this.lexer.inlineQueue[e].src)) {
						this.lexer.inlineQueue[e].src = this.lexer.inlineQueue[e].src.replace(this.rules.other.listReplaceTask, "");
						break;
					}
					let n = this.rules.other.listTaskCheckbox.exec(e.raw);
					if (n) {
						let t = {
							type: "checkbox",
							raw: n[0] + " ",
							checked: n[0] !== "[ ]"
						};
						e.checked = t.checked, o.loose ? e.tokens[0] && ["paragraph", "text"].includes(e.tokens[0].type) && "tokens" in e.tokens[0] && e.tokens[0].tokens ? (e.tokens[0].raw = t.raw + e.tokens[0].raw, e.tokens[0].text = t.raw + e.tokens[0].text, e.tokens[0].tokens.unshift(t)) : e.tokens.unshift({
							type: "paragraph",
							raw: t.raw,
							text: t.raw,
							tokens: [t]
						}) : e.tokens.unshift(t);
					}
				} else e.task &&= !1;
			}
			if (o.loose) for (let e of o.items) {
				e.loose = !0;
				for (let t of e.tokens) t.type === "text" && (t.type = "paragraph");
			}
			return o;
		}
	}
	html(e) {
		let t = this.rules.block.html.exec(e);
		if (t) {
			let e = Ja(t[0]);
			return {
				type: "html",
				block: !0,
				raw: e,
				pre: t[1] === "pre" || t[1] === "script" || t[1] === "style",
				text: e
			};
		}
	}
	def(e) {
		let t = this.rules.block.def.exec(e);
		if (t) {
			if (!this.rules.other.startAngleBracket.test(t[2]) && Xa(t[2], "()") !== -1) return;
			let e = Ya(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: qa(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Ka(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: qa(t[0], "\n"),
			header: [],
			align: [],
			rows: []
		};
		if (n.length === r.length) {
			for (let e of r) this.rules.other.tableAlignRight.test(e) ? a.align.push("right") : this.rules.other.tableAlignCenter.test(e) ? a.align.push("center") : this.rules.other.tableAlignLeft.test(e) ? a.align.push("left") : a.align.push(null);
			for (let e = 0; e < n.length; e++) a.header.push({
				text: n[e],
				tokens: this.lexer.inline(n[e]),
				header: !0,
				align: a.align[e]
			});
			for (let e of i) a.rows.push(Ka(e, a.header.length).map((e, t) => ({
				text: e,
				tokens: this.lexer.inline(e),
				header: !1,
				align: a.align[t]
			})));
			return a;
		}
	}
	lheading(e) {
		let t = this.rules.block.lheading.exec(e);
		if (t) {
			let e = t[1].trim();
			return {
				type: "heading",
				raw: qa(t[0], "\n"),
				depth: t[2].charAt(0) === "=" ? 1 : 2,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	paragraph(e) {
		let t = this.rules.block.paragraph.exec(e);
		if (t) {
			let e = t[1].charAt(t[1].length - 1) === "\n" ? t[1].slice(0, -1) : t[1];
			return {
				type: "paragraph",
				raw: t[0],
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	text(e) {
		let t = this.rules.block.text.exec(e);
		if (t) return {
			type: "text",
			raw: t[0],
			text: t[0],
			tokens: this.lexer.inline(t[0])
		};
	}
	escape(e) {
		let t = this.rules.inline.escape.exec(e);
		if (t) return {
			type: "escape",
			raw: t[0],
			text: t[1]
		};
	}
	tag(e) {
		let t = this.rules.inline.tag.exec(e);
		if (t) return !this.lexer.state.inLink && this.rules.other.startATag.test(t[0]) ? this.lexer.state.inLink = !0 : this.lexer.state.inLink && this.rules.other.endATag.test(t[0]) && (this.lexer.state.inLink = !1), !this.lexer.state.inRawBlock && this.rules.other.startPreScriptTag.test(t[0]) ? this.lexer.state.inRawBlock = !0 : this.lexer.state.inRawBlock && this.rules.other.endPreScriptTag.test(t[0]) && (this.lexer.state.inRawBlock = !1), {
			type: "html",
			raw: t[0],
			inLink: this.lexer.state.inLink,
			inRawBlock: this.lexer.state.inRawBlock,
			block: !1,
			text: t[0]
		};
	}
	link(e) {
		if (this.lexer.state.linkParenPossible === !1) return;
		let t = this.rules.inline.link.exec(e);
		if (t) {
			let n = t[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && oo(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = qa(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Xa(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = qa(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), io(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && oo(e, n[1], r, this.rules)) return;
			let i = t[Ya(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return io(n, i, n[0], this.lexer, this.rules);
		}
	}
	emStrong(e, t, n = "") {
		let r = this.rules.inline.emStrongLDelim.exec(e);
		if (!r || !r[1] && !r[2] && !r[3] && !r[4]) return;
		let i = r[0][0], a = n === i, o = t.length - e.length, s = o;
		for (; !this.options.pedantic && s > 0 && t[s - 1] === i;) s--;
		let c = [...t.slice(Math.max(0, s - 2), s)].at(-1) ?? "", l = c === "*" || c === "_" || !!this.rules.inline.punctuation.exec(c);
		s < o && (n = c);
		let u = s < o ? l : !!this.rules.inline.punctuation.exec(n);
		if (r[4] && n.match(this.rules.other.unicodeAlphaNumeric)) return;
		let d = r[1] || r[3] || "";
		if (!d || !n || u) {
			let n = [...r[0]].length - 1, u, f, p = n, m = 0, h = o - s + n, g = !this.options.pedantic && !!c.trim() && (!!d || i === "*" && !l);
			if (i !== "*" && i !== "_") return;
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = to(this, t, i), y = t.length - e.length + n;
			if (v < y) return;
			t = t.slice(-1 * e.length + n);
			let b = v - y;
			for (_.lastIndex = 0; (r = _.exec(t)) !== null && !(r.index > b);) {
				if (u = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !u) continue;
				if (f = [...u].length, r[3] || r[4]) {
					p += f;
					continue;
				}
				if (r[5] || r[6] || g) {
					if (a && (r[5] || r[6])) break;
					if (h % 3 && !((h + f) % 3)) {
						m += f;
						continue;
					}
				}
				if (p -= f, p > 0) continue;
				f = Math.min(f, f + p + m);
				let t = [...r[0]][0].length, i = e.slice(0, n + r.index + t + f);
				if (Math.min(n, f) % 2) {
					let e = i.slice(1, -1);
					return {
						type: "em",
						raw: i,
						text: e,
						tokens: this.lexer.inlineTokens(e)
					};
				}
				let o = i.slice(2, -2);
				return {
					type: "strong",
					raw: i,
					text: o,
					tokens: this.lexer.inlineTokens(o)
				};
			}
		}
	}
	codespan(e) {
		let t = this.rules.inline.code.exec(e);
		if (t) {
			let e = t[2].replace(this.rules.other.newLineCharGlobal, " "), n = this.rules.other.nonSpaceChar.test(e), r = this.rules.other.startingSpaceChar.test(e) && this.rules.other.endingSpaceChar.test(e);
			return n && r && (e = e.substring(1, e.length - 1)), {
				type: "codespan",
				raw: t[0],
				text: e
			};
		}
	}
	br(e) {
		let t = this.rules.inline.br.exec(e);
		if (t) return {
			type: "br",
			raw: t[0]
		};
	}
	del(e, t, n = "") {
		let r = this.rules.inline.delLDelim.exec(e);
		if (r && (!r[1] || !n || this.rules.inline.punctuation.exec(n))) {
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = no(this, t), l = t.length - e.length + n;
			if (c < l) return;
			t = t.slice(-1 * e.length + n);
			let u = c - l;
			for (s.lastIndex = 0; (r = s.exec(t)) !== null && !(r.index > u);) {
				if (i = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !i || (a = [...i].length, a !== n)) continue;
				if (r[3] || r[4]) {
					o += a;
					continue;
				}
				if (o -= a, o > 0) continue;
				a = Math.min(a, a + o);
				let t = [...r[0]][0].length, s = e.slice(0, n + r.index + t + a), c = s.slice(n, -n);
				return {
					type: "del",
					raw: s,
					text: c,
					tokens: this.lexer.inlineTokens(c)
				};
			}
		}
	}
	autolink(e) {
		let t = this.rules.inline.autolink.exec(e);
		if (t) {
			let e, n;
			return t[2] === "@" ? (e = t[1], n = "mailto:" + e) : (e = t[1], n = e), {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				autolink: !0,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	url(e) {
		let t;
		if (t = this.rules.inline.url.exec(e)) {
			let e, n;
			if (t[2] === "@") e = t[0], n = "mailto:" + e;
			else {
				let r;
				do
					r = t[0], t[0] = this.rules.inline._backpedal.exec(t[0])?.[0] ?? "";
				while (r !== t[0]);
				e = t[0], n = t[1] === "www." ? "http://" + t[0] : t[0];
			}
			return {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				autolink: !0,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	inlineText(e) {
		let t = this.rules.inline.text.exec(e);
		if (t) {
			let e = this.lexer.state.inRawBlock;
			return {
				type: "text",
				raw: t[0],
				text: e ? t[0] : Wa(t[0]),
				escaped: e
			};
		}
	}
}, co = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || bi, this.options.tokenizer = this.options.tokenizer || new so(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: Y,
			block: za.normal,
			inline: Ba.normal
		};
		this.options.pedantic ? (t.block = za.pedantic, t.inline = Ba.pedantic) : this.options.gfm && (t.block = za.gfm, t.inline = this.options.breaks ? Ba.breaks : Ba.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: za,
			inline: Ba
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(Y.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(Y.tabCharGlobal, "    ").replace(Y.spaceLine, ""));
		let r = 1 / 0;
		for (; e;) {
			if (e.length < r) r = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			let i;
			if (this.options.extensions?.block?.some((n) => (i = n.call({ lexer: this }, e, t)) ? (e = e.substring(i.raw.length), t.push(i), !0) : !1)) continue;
			if (i = this.tokenizer.space(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				i.raw.length === 1 && n !== void 0 ? n.raw += "\n" : t.push(i);
				continue;
			}
			if (i = this.tokenizer.code(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.at(-1).src = n.text) : t.push(i);
				continue;
			}
			if ((i = this.tokenizer.fences(e)) || (i = this.tokenizer.heading(e)) || (i = this.tokenizer.hr(e)) || (i = this.tokenizer.blockquote(e)) || (i = this.tokenizer.list(e)) || (i = this.tokenizer.html(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			if (i = this.tokenizer.def(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.raw, this.inlineQueue.at(-1).src = n.text) : (this.tokens.links[i.tag] || (this.tokens.links[i.tag] = {
					href: i.href,
					title: i.title
				}), t.push(i));
				continue;
			}
			if ((i = this.tokenizer.table(e)) || (i = this.tokenizer.lheading(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			let a = e;
			if (this.options.extensions?.startBlock) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startBlock.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (a = e.substring(0, t + 1));
			}
			if (this.state.top && (i = this.tokenizer.paragraph(a))) {
				let r = t.at(-1);
				n && r?.type === "paragraph" ? (r.raw += (r.raw.endsWith("\n") ? "" : "\n") + i.raw, r.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = r.text) : t.push(i), n = a.length !== e.length, e = e.substring(i.raw.length);
			} else if (i = this.tokenizer.text(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = n.text) : t.push(i);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return this.state.top = !0, t;
	}
	inline(e, t = []) {
		return this.inlineQueue.push({
			src: e,
			tokens: t
		}), t;
	}
	linkInText(e) {
		if (!e.includes("[")) return !1;
		let t = this.tokenizer.rules.inline.link;
		for (let n of e.matchAll(this.tokenizer.rules.inline.blockSkip)) if (t.test(n[0]) && e.charAt(n.index - 1) !== "!") return !0;
		for (let t of e.matchAll(this.tokenizer.rules.inline.reflinkSearch)) {
			let e = t[0], n = e.lastIndexOf("[");
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Ya(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
		}
		return !1;
	}
	inlineTokens(e, t = []) {
		this.tokenizer.lexer = this;
		let n = this.state.linkParenPossible;
		this.state.linkParenPossible = n && e.includes(")");
		try {
			return this.#e(e, t);
		} finally {
			this.state.linkParenPossible = n;
		}
	}
	#e(e, t) {
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== so.prototype.emStrong || this.tokenizer.del !== so.prototype.del ? this.#n(e) : e;
		$a(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			eo(this.tokenizer);
		}
	}
	#t(e, t, n) {
		let r = !1, i = "", a = 1 / 0;
		for (; e;) {
			if (e.length < a) a = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			r || (i = ""), r = !1;
			let o;
			if (this.options.extensions?.inline?.some((t) => (o = t.call({ lexer: this }, e, n)) ? (e = e.substring(o.raw.length), n.push(o), !0) : !1)) continue;
			if ((o = this.tokenizer.escape(e)) || (o = this.tokenizer.tag(e)) || (o = this.tokenizer.link(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			if (o = this.tokenizer.reflink(e, this.tokens.links)) {
				e = e.substring(o.raw.length);
				let t = n.at(-1);
				o.type === "text" && t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
				continue;
			}
			if ((o = this.tokenizer.emStrong(e, t, i)) || (o = this.tokenizer.codespan(e)) || (o = this.tokenizer.br(e)) || (o = this.tokenizer.del(e, t, i)) || (o = this.tokenizer.autolink(e)) || !this.state.inLink && (o = this.tokenizer.url(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			let s = e;
			if (this.options.extensions?.startInline) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startInline.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (s = e.substring(0, t + 1));
			}
			if (o = this.tokenizer.inlineText(s)) {
				e = e.substring(o.raw.length), o.raw.slice(-1) !== "_" && (i = o.raw.slice(-1)), r = !0;
				let t = n.at(-1);
				t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return n;
	}
	#n(e) {
		let t = e;
		if (this.tokens.links && e.includes("[")) {
			let e = this.tokenizer.rules.inline.reflinkSearch, n = (t) => {
				let r = t.lastIndexOf("[");
				if (!Object.hasOwn(this.tokens.links, Ya(t.slice(r + 1, -1)))) return t;
				if (r > 1 && t.charAt(0) !== "!") {
					let i = t.slice(1, r - 1);
					if (this.linkInText(i)) return "[" + i.replace(e, n) + "][" + "a".repeat(t.length - r - 2) + "]";
				}
				return "[" + "a".repeat(t.length - 2) + "]";
			};
			t = t.replace(e, n);
		}
		return t = t.replace(this.tokenizer.rules.inline.anyPunctuationOutsideCode, (e, t, n) => n ? "+".repeat(e.length) : t ? t + e.slice(t.length, -t.length).replace(/`/g, "+") + t : e), t = t.replace(this.tokenizer.rules.inline.blockSkip, (e, t, n) => {
			let r = n ? n.length : 0;
			return e.slice(0, r) + "[" + "a".repeat(e.length - r - 2) + "]";
		}), t = this.options.hooks?.emStrongMask?.call({ lexer: this }, t) ?? t, t;
	}
	infiniteLoopError(e) {
		let t = "Infinite loop on byte: " + e;
		if (this.options.silent) console.error(t);
		else throw Error(t);
	}
}, lo = class {
	options;
	parser;
	constructor(e) {
		this.options = e || bi;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(Y.notSpaceStart)?.[0], i = e ? e.replace(Y.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Ua(r) + "\">" + (n ? i : Ua(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Ua(i, !0)) + "</code></pre>\n";
	}
	blockquote({ tokens: e }) {
		return `<blockquote>
${this.parser.parse(e)}</blockquote>
`;
	}
	html({ text: e }) {
		return e;
	}
	def(e) {
		return "";
	}
	heading({ tokens: e, depth: t }) {
		return `<h${t}>${this.parser.parseInline(e)}</h${t}>
`;
	}
	hr(e) {
		return "<hr>\n";
	}
	list(e) {
		let t = e.ordered, n = e.start, r = "";
		for (let t = 0; t < e.items.length; t++) {
			let n = e.items[t];
			r += this.listitem(n);
		}
		let i = t ? "ol" : "ul", a = t && n !== 1 ? " start=\"" + n + "\"" : "";
		return "<" + i + a + ">\n" + r + "</" + i + ">\n";
	}
	listitem(e) {
		return `<li>${this.parser.parse(e.tokens)}</li>
`;
	}
	checkbox({ checked: e }) {
		return "<input " + (e ? "checked=\"\" " : "") + "disabled=\"\" type=\"checkbox\"> ";
	}
	paragraph({ tokens: e }) {
		return `<p>${this.parser.parseInline(e)}</p>
`;
	}
	table(e) {
		let t = "", n = "";
		for (let t = 0; t < e.header.length; t++) n += this.tablecell(e.header[t]);
		t += this.tablerow({ text: n });
		let r = "";
		for (let t = 0; t < e.rows.length; t++) {
			let i = e.rows[t];
			n = "";
			for (let e = 0; e < i.length; e++) n += this.tablecell(i[e]);
			r += this.tablerow({ text: n });
		}
		return r &&= `<tbody>${r}</tbody>`, "<table>\n<thead>\n" + t + "</thead>\n" + r + "</table>\n";
	}
	tablerow({ text: e }) {
		return `<tr>
${e}</tr>
`;
	}
	tablecell(e) {
		let t = this.parser.parseInline(e.tokens), n = e.header ? "th" : "td";
		return (e.align ? `<${n} align="${e.align}">` : `<${n}>`) + t + `</${n}>
`;
	}
	strong({ tokens: e }) {
		return `<strong>${this.parser.parseInline(e)}</strong>`;
	}
	em({ tokens: e }) {
		return `<em>${this.parser.parseInline(e)}</em>`;
	}
	codespan({ text: e }) {
		return `<code>${Ua(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Ua(n, !0) : this.parser.parseInline(r), o = Ga(e);
		if (o === null) return a;
		e = Ua(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Ua(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Ga(e);
		if (i === null) return Ua(n);
		e = i;
		let a = `<img src="${Ua(e)}" alt="${Ua(n)}"`;
		return t && (a += ` title="${Ua(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Ua(e.text);
	}
}, uo = class {
	parser;
	strong({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	em({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	codespan({ text: e }) {
		return e;
	}
	del({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	html({ text: e }) {
		return e;
	}
	text({ text: e }) {
		return e;
	}
	link({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	image({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	br() {
		return "";
	}
	checkbox({ raw: e }) {
		return e;
	}
}, fo = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || bi, this.options.renderer = this.options.renderer || new lo(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new uo(), this.textRenderer.parser = this;
	}
	static parse(t, n) {
		return new e(n).parse(t);
	}
	static parseInline(t, n) {
		return new e(n).parseInline(t);
	}
	parse(e) {
		this.renderer.parser = this;
		let t = "";
		for (let n = 0; n < e.length; n++) {
			let r = e[n];
			if (this.options.extensions?.renderers?.[r.type]) {
				let e = r, n = this.options.extensions.renderers[e.type].call({ parser: this }, e);
				if (n !== !1 || ![
					"space",
					"hr",
					"heading",
					"code",
					"table",
					"blockquote",
					"list",
					"checkbox",
					"html",
					"def",
					"paragraph",
					"text"
				].includes(e.type)) {
					t += n || "";
					continue;
				}
			}
			let i = r;
			switch (i.type) {
				case "space":
					t += this.renderer.space(i);
					break;
				case "hr":
					t += this.renderer.hr(i);
					break;
				case "heading":
					t += this.renderer.heading(i);
					break;
				case "code":
					t += this.renderer.code(i);
					break;
				case "table":
					t += this.renderer.table(i);
					break;
				case "blockquote":
					t += this.renderer.blockquote(i);
					break;
				case "list":
					t += this.renderer.list(i);
					break;
				case "checkbox":
					t += this.renderer.checkbox(i);
					break;
				case "html":
					t += this.renderer.html(i);
					break;
				case "def":
					t += this.renderer.def(i);
					break;
				case "paragraph":
					t += this.renderer.paragraph(i);
					break;
				case "text":
					t += this.renderer.text(i);
					break;
				default: {
					let e = "Token with \"" + i.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return t;
	}
	parseInline(e, t = this.renderer) {
		this.renderer.parser = this;
		let n = "";
		for (let r = 0; r < e.length; r++) {
			let i = e[r];
			if (this.options.extensions?.renderers?.[i.type]) {
				let e = this.options.extensions.renderers[i.type].call({ parser: this }, i);
				if (e !== !1 || ![
					"escape",
					"html",
					"link",
					"image",
					"checkbox",
					"strong",
					"em",
					"codespan",
					"br",
					"del",
					"text"
				].includes(i.type)) {
					n += e || "";
					continue;
				}
			}
			let a = i;
			switch (a.type) {
				case "escape":
					n += t.text(a);
					break;
				case "html":
					n += t.html(a);
					break;
				case "link":
					n += t.link(a);
					break;
				case "image":
					n += t.image(a);
					break;
				case "checkbox":
					n += t.checkbox(a);
					break;
				case "strong":
					n += t.strong(a);
					break;
				case "em":
					n += t.em(a);
					break;
				case "codespan":
					n += t.codespan(a);
					break;
				case "br":
					n += t.br(a);
					break;
				case "del":
					n += t.del(a);
					break;
				case "text":
					n += t.text(a);
					break;
				default: {
					let e = "Token with \"" + a.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return n;
	}
}, po = class {
	options;
	block;
	constructor(e) {
		this.options = e || bi;
	}
	static passThroughHooks = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens",
		"emStrongMask"
	]);
	static passThroughHooksRespectAsync = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens"
	]);
	preprocess(e) {
		return e;
	}
	postprocess(e) {
		return e;
	}
	processAllTokens(e) {
		return e;
	}
	emStrongMask(e) {
		return e;
	}
	provideLexer(e = this.block) {
		return e ? co.lex : co.lexInline;
	}
	provideParser(e = this.block) {
		return e ? fo.parse : fo.parseInline;
	}
};
function mo(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var ho = new class {
	defaults = yi();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = fo;
	Renderer = lo;
	TextRenderer = uo;
	Lexer = co;
	Tokenizer = so;
	Hooks = po;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (mo(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) mo(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) mo(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				mo(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					mo(n, this.walkTokens(i, t));
				}) : e.tokens && mo(n, this.walkTokens(e.tokens, t));
			}
		}
		return n;
	}
	use(...e) {
		let t = this.defaults.extensions || {
			renderers: Object.create(null),
			childTokens: Object.create(null)
		};
		return e.forEach((e) => {
			let n = { ...e };
			if (n.async = this.defaults.async || n.async || !1, e.extensions && (e.extensions.forEach((e) => {
				if (!e.name) throw Error("extension name required");
				if ("renderer" in e) {
					let n = t.renderers[e.name];
					n ? t.renderers[e.name] = function(...t) {
						let r = e.renderer.apply(this, t);
						return r === !1 && (r = n.apply(this, t)), r;
					} : t.renderers[e.name] = e.renderer;
				}
				if ("tokenizer" in e) {
					if (!e.level || e.level !== "block" && e.level !== "inline") throw Error("extension level must be 'block' or 'inline'");
					let n = t[e.level];
					n ? n.unshift(e.tokenizer) : t[e.level] = [e.tokenizer], e.start && (e.level === "block" ? t.startBlock ? t.startBlock.push(e.start) : t.startBlock = [e.start] : e.level === "inline" && (t.startInline ? t.startInline.push(e.start) : t.startInline = [e.start]));
				}
				"childTokens" in e && e.childTokens && (t.childTokens[e.name] = e.childTokens);
			}), n.extensions = t), e.renderer) {
				let t = this.defaults.renderer || new lo(this.defaults);
				for (let n in e.renderer) {
					if (!(n in t)) throw Error(`renderer '${n}' does not exist`);
					if (["options", "parser"].includes(n)) continue;
					let r = n, i = e.renderer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n === !1 ? "" : n ?? "";
					};
				}
				n.renderer = t;
			}
			if (e.tokenizer) {
				let t = this.defaults.tokenizer || new so(this.defaults);
				for (let n in e.tokenizer) {
					if (!(n in t)) throw Error(`tokenizer '${n}' does not exist`);
					if ([
						"options",
						"rules",
						"lexer"
					].includes(n)) continue;
					let r = n, i = e.tokenizer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.tokenizer = t;
			}
			if (e.hooks) {
				let t = this.defaults.hooks || new po();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = po.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && po.passThroughHooksRespectAsync.has(n)) return (async () => {
							let n = await i.call(t, e);
							return a.call(t, n);
						})();
						let r = i.call(t, e);
						return a.call(t, r);
					} : (...e) => {
						if (t.options.async) return (async () => {
							let n = await i.apply(t, e);
							return n === !1 && (n = await a.apply(t, e)), n;
						})();
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.hooks = t;
			}
			if (e.walkTokens) {
				let t = this.defaults.walkTokens, r = e.walkTokens;
				n.walkTokens = function(e) {
					let n = [];
					return mo(n, r.call(this, e)), t && mo(n, t.call(this, e)), n;
				};
			}
			this.defaults = {
				...this.defaults,
				...n
			};
		}), this;
	}
	setOptions(e) {
		return this.defaults = {
			...this.defaults,
			...e
		}, this;
	}
	lexer(e, t) {
		return co.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return fo.parse(e, t ?? this.defaults);
	}
	parseMarkdown(e) {
		return (t, n) => {
			let r = { ...n }, i = {
				...this.defaults,
				...r
			}, a = this.onError(!!i.silent, !!i.async);
			if (this.defaults.async === !0 && r.async === !1) return a(/* @__PURE__ */ Error("marked(): The async option was set to true by an extension. Remove async: false from the parse options object to return a Promise."));
			if (typeof t > "u" || t === null) return a(/* @__PURE__ */ Error("marked(): input parameter is undefined or null"));
			if (typeof t != "string") return a(/* @__PURE__ */ Error("marked(): input parameter is of type " + Object.prototype.toString.call(t) + ", string expected"));
			if (i.hooks && (i.hooks.options = i, i.hooks.block = e), i.async) return (async () => {
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? co.lex : co.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? fo.parse : fo.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? co.lex : co.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? fo.parse : fo.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Ua(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return ho.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return ho.setOptions(e), X.defaults = ho.defaults, xi(X.defaults), X;
}, X.getDefaults = yi, X.defaults = bi;
function go(...e) {
	return ho.use(...e), X.defaults = ho.defaults, xi(X.defaults), X;
}
X.use = go, X.walkTokens = function(e, t) {
	return ho.walkTokens(e, t);
}, X.parseInline = ho.parseInline, X.Parser = fo, X.parser = fo.parse, X.Renderer = lo, X.TextRenderer = uo, X.Lexer = co, X.lexer = co.lex, X.Tokenizer = so, X.Hooks = po, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, fo.parse, co.lex;
// node_modules/dompurify/dist/purify.es.mjs
function _o(e, t) {
	this.v = e, this.k = t;
}
function vo(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function yo(e) {
	if (Array.isArray(e)) return e;
}
function bo(e, t) {
	var n = e == null ? null : typeof Symbol < "u" && e[Symbol.iterator] || e["@@iterator"];
	if (n != null) {
		var r, i, a, o, s = [], c = !0, l = !1;
		try {
			if (a = (n = n.call(e)).next, t === 0) {
				if (Object(n) !== n) return;
				c = !1;
			} else for (; !(c = (r = a.call(n)).done) && (s.push(r.value), s.length !== t); c = !0);
		} catch (e) {
			l = !0, i = e;
		} finally {
			try {
				if (!c && n.return != null && (o = n.return(), Object(o) !== o)) return;
			} finally {
				if (l) throw i;
			}
		}
		return s;
	}
}
function xo() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function So(e, t) {
	return yo(e) || bo(e, t) || Co(e, t) || xo();
}
function Co(e, t) {
	if (e) {
		if (typeof e == "string") return vo(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? vo(e, t) : void 0;
	}
}
function wo(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof _o;
			Promise.resolve(s ? o.v : o).then(function(n) {
				if (s) {
					var c = t === "return" && o.k ? t : "next";
					if (!o.k || n.done) return r(c, n);
					n = e[c](n).value;
				}
				i(!!a.done, n);
			}, function(e) {
				r("throw", e);
			});
		} catch (e) {
			i(2, e);
		}
	}
	function i(e, i) {
		e === 2 ? t.reject(i) : t.resolve({
			value: i,
			done: e
		}), (t = t.next) ? r(t.key, t.arg) : n = null;
	}
	this._invoke = function(e, i) {
		return new Promise(function(a, o) {
			var s = {
				key: e,
				arg: i,
				resolve: a,
				reject: o,
				next: null
			};
			n ? n = n.next = s : (t = n = s, r(e, i));
		});
	}, typeof e.return != "function" && (this.return = void 0);
}
wo.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, wo.prototype.next = function(e) {
	return this._invoke("next", e);
}, wo.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, wo.prototype.return = function(e) {
	return this._invoke("return", e);
};
var To = Object.entries, Eo = Object.setPrototypeOf, Do = Object.isFrozen, Oo = Object.getPrototypeOf, ko = Object.getOwnPropertyDescriptor, Ao = Object.freeze, jo = Object.seal, Mo = Object.create, No = typeof Reflect < "u" && Reflect, Po = No.apply, Fo = No.construct;
Ao ||= function(e) {
	return e;
}, jo ||= function(e) {
	return e;
}, Po ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Fo ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Io = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var Lo = Z(Array.prototype.lastIndexOf), Ro = Z(Array.prototype.pop), zo = Z(Array.prototype.push), Bo = Z(Array.prototype.slice), Vo = Z(Array.prototype.splice), Ho = Array.isArray, Uo = os(Set), Wo = Z(Set.prototype.add), Go = Z(Set.prototype.has), Ko = Z(String.prototype.toLowerCase), qo = Z(String.prototype.toString), Jo = Z(String.prototype.match), Yo = Z(String.prototype.replace), Xo = Z(String.prototype.indexOf), Zo = Z(String.prototype.trim), Qo = Z(Number.prototype.toString), $o = Z(Boolean.prototype.toString), es = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), ts = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), ns = Z(Object.prototype.hasOwnProperty), rs = Z(Object.prototype.toString), is = Z(RegExp.prototype.test), as = os(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Po(e, t, n);
	};
}
function os(e) {
	return function() {
		return Fo(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Ko;
	if (Eo && Eo(e, null), !Ho(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Do(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function ss(e) {
	for (let t = 0; t < e.length; t++) ns(e, t) || (e[t] = null);
	return e;
}
function cs(e) {
	let t = Mo(null);
	for (let r of To(e)) {
		var n = So(r, 2);
		let i = n[0], a = n[1];
		ns(e, i) && (t[i] = Ho(a) ? ss(a) : a && typeof a == "object" && a.constructor === Object ? cs(a) : a);
	}
	return t;
}
function ls(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return Qo(e);
		case "boolean": return $o(e);
		case "bigint": return es ? es(e) : "0";
		case "symbol": return ts ? ts(e) : "Symbol()";
		case "undefined": return rs(e);
		case "function":
		case "object": {
			if (e === null) return rs(e);
			let t = e, n = us(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : rs(e);
			}
			return rs(e);
		}
		default: return rs(e);
	}
}
function us(e, t) {
	for (; e !== null;) {
		let n = ko(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Oo(e);
	}
	function n() {
		return null;
	}
	return n;
}
function ds(e, t) {
	for (; e !== null;) {
		let n = ko(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = Oo(e);
	}
	return null;
}
function fs(e) {
	try {
		return is(e, ""), !0;
	} catch {
		return !1;
	}
}
var ps = Ao(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), ms = Ao(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), hs = Ao([
	"feBlend",
	"feColorMatrix",
	"feComponentTransfer",
	"feComposite",
	"feConvolveMatrix",
	"feDiffuseLighting",
	"feDisplacementMap",
	"feDistantLight",
	"feDropShadow",
	"feFlood",
	"feFuncA",
	"feFuncB",
	"feFuncG",
	"feFuncR",
	"feGaussianBlur",
	"feImage",
	"feMerge",
	"feMergeNode",
	"feMorphology",
	"feOffset",
	"fePointLight",
	"feSpecularLighting",
	"feSpotLight",
	"feTile",
	"feTurbulence"
]), gs = Ao([
	"animate",
	"color-profile",
	"cursor",
	"discard",
	"font-face",
	"font-face-format",
	"font-face-name",
	"font-face-src",
	"font-face-uri",
	"foreignobject",
	"hatch",
	"hatchpath",
	"mesh",
	"meshgradient",
	"meshpatch",
	"meshrow",
	"missing-glyph",
	"script",
	"set",
	"solidcolor",
	"unknown",
	"use"
]), _s = Ao(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), vs = Ao([
	"maction",
	"maligngroup",
	"malignmark",
	"mlongdiv",
	"mscarries",
	"mscarry",
	"msgroup",
	"mstack",
	"msline",
	"msrow",
	"semantics",
	"annotation",
	"annotation-xml",
	"mprescripts",
	"none"
]), ys = Ao(["#text"]), bs = Ao(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), xs = Ao(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), Ss = Ao(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), Cs = Ao([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), ws = jo(/{{[\w\W]*|^[\w\W]*}}/g), Ts = jo(/<%[\w\W]*|^[\w\W]*%>/g), Es = jo(/\${[\w\W]*/g), Ds = jo(/^data-[\-\w.\u00B7-\uFFFF]+$/), Os = jo(/^aria-[\-\w]+$/), ks = jo(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), As = jo(/^(?:\w+script|data):/i), js = jo(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Ms = jo(/^html$/i), Ns = jo(/^[a-z][.\w]*(-[.\w]+)+$/i), Ps = jo(/<[/\w!]/g), Fs = jo(/<[/\w]/g), Is = jo(/<\/no(script|embed|frames)/i), Ls = jo(/\/>/i), $ = {
	element: 1,
	attribute: 2,
	text: 3,
	cdataSection: 4,
	entityReference: 5,
	entityNode: 6,
	processingInstruction: 7,
	comment: 8,
	document: 9,
	documentType: 10,
	documentFragment: 11,
	notation: 12
}, Rs = Ao(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), zs = function() {
	return typeof window > "u" ? null : window;
}, Bs = function(e, t) {
	if (typeof e != "object" || typeof e.createPolicy != "function") return null;
	let n = null, r = "data-tt-policy-suffix";
	t && t.hasAttribute(r) && (n = t.getAttribute(r));
	let i = "dompurify" + (n ? "#" + n : "");
	try {
		return e.createPolicy(i, {
			createHTML(e) {
				return e;
			},
			createScriptURL(e) {
				return e;
			}
		});
	} catch {
		return console.warn("TrustedTypes policy " + i + " could not be created."), null;
	}
}, Vs = function() {
	return {
		afterSanitizeAttributes: [],
		afterSanitizeElements: [],
		afterSanitizeShadowDOM: [],
		beforeSanitizeAttributes: [],
		beforeSanitizeElements: [],
		beforeSanitizeShadowDOM: [],
		uponSanitizeAttribute: [],
		uponSanitizeElement: [],
		uponSanitizeShadowNode: []
	};
}, Hs = function(e, t, n, r) {
	return ns(e, t) && Ho(e[t]) ? Q(r.base ? cs(r.base) : {}, e[t], r.transform) : n;
}, Us = function(e, t, n) {
	let r = ns(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? cs(r) : n();
};
function Ws() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : zs(), t = (e) => Ws(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = us(d, "cloneNode"), p = us(d, "remove"), m = us(d, "removeAttributeNode"), h = us(d, "nextSibling"), g = us(d, "childNodes"), _ = us(d, "parentNode"), v = us(d, "shadowRoot"), y = us(d, "attributes"), b = o && o.prototype ? us(o.prototype, "nodeType") : null, x = o && o.prototype ? us(o.prototype, "nodeName") : null, S = o && o.prototype ? us(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = ds(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, ne = "", re, ie = !1, ae = 0, oe = function() {
		if (ae > 0) throw as("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, se = function(e) {
		oe(), ae++;
		try {
			return T.createHTML(e);
		} finally {
			ae--;
		}
	}, ce = function(e) {
		oe(), ae++;
		try {
			return T.createScriptURL(e);
		} finally {
			ae--;
		}
	}, le = function() {
		return ie ||= (re = Bs(u, i), !0), re;
	}, ue = n, de = ue.implementation, fe = ue.createNodeIterator, pe = ue.createDocumentFragment, me = ue.getElementsByTagName, he = r.importNode, E = Vs();
	t.isSupported = typeof To == "function" && typeof _ == "function" && de && de.createHTMLDocument !== void 0;
	let ge = ws, _e = Ts, ve = Es, ye = Ds, be = Os, D = As, xe = js, O = Ns, Se = ks, k = null, A = Q({}, [
		...ps,
		...ms,
		...hs,
		..._s,
		...ys
	]), j = null, Ce = Q({}, [
		...bs,
		...xs,
		...Ss,
		...Cs
	]), we = Object.seal(Mo(null, {
		tagNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeNameCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		allowCustomizedBuiltInElements: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: !1
		}
	})), Te = null, Ee = null, De = Object.seal(Mo(null, {
		tagCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		},
		attributeCheck: {
			writable: !0,
			configurable: !1,
			enumerable: !0,
			value: null
		}
	})), Oe = !0, ke = !0, Ae = !1, je = !0, Me = !1, Ne = !0, Pe = !1, Fe = !1, Ie = null, Le = null, Re = !1, ze = !1, Be = !1, Ve = !1, He = !0, Ue = !1, We = "user-content-", Ge = !0, Ke = !1, qe = null, Je = {}, Ye = null, Xe = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ze = null, Qe = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), M = null, $e = Q({}, [
		"alt",
		"class",
		"for",
		"id",
		"label",
		"name",
		"pattern",
		"placeholder",
		"role",
		"summary",
		"title",
		"value",
		"style",
		"xmlns"
	]), et = "http://www.w3.org/1998/Math/MathML", tt = "http://www.w3.org/2000/svg", nt = "http://www.w3.org/1999/xhtml", rt = nt, it = !1, at = null, ot = Q({}, [
		et,
		tt,
		nt
	], qo), st = Ao([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), ct = Q({}, st), lt = Ao(["annotation-xml"]), ut = Q({}, lt), dt = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), ft = null, pt = ["application/xhtml+xml", "text/html"], N = null, mt = null, ht = n.createElement("form"), gt = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, _t = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (mt && mt === e) return;
		e && typeof e == "object" || (e = {}), e = cs(e), ft = pt.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, N = ft === "application/xhtml+xml" ? qo : Ko, k = Hs(e, "ALLOWED_TAGS", A, { transform: N }), j = Hs(e, "ALLOWED_ATTR", Ce, { transform: N }), at = Hs(e, "ALLOWED_NAMESPACES", ot, { transform: qo }), M = Hs(e, "ADD_URI_SAFE_ATTR", $e, {
			transform: N,
			base: $e
		}), Ze = Hs(e, "ADD_DATA_URI_TAGS", Qe, {
			transform: N,
			base: Qe
		}), Ye = Hs(e, "FORBID_CONTENTS", Xe, { transform: N }), Te = Hs(e, "FORBID_TAGS", cs({}), { transform: N }), Ee = Hs(e, "FORBID_ATTR", cs({}), { transform: N }), Je = ns(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? cs(e.USE_PROFILES) : e.USE_PROFILES : !1, Oe = e.ALLOW_ARIA_ATTR !== !1, ke = e.ALLOW_DATA_ATTR !== !1, Ae = e.ALLOW_UNKNOWN_PROTOCOLS || !1, je = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Me = e.SAFE_FOR_TEMPLATES || !1, Ne = e.SAFE_FOR_XML !== !1, Pe = e.WHOLE_DOCUMENT || !1, ze = e.RETURN_DOM || !1, Be = e.RETURN_DOM_FRAGMENT || !1, Ve = e.RETURN_TRUSTED_TYPE || !1, Re = e.FORCE_BODY || !1, He = e.SANITIZE_DOM !== !1, Ue = e.SANITIZE_NAMED_PROPS || !1, Ge = e.KEEP_CONTENT !== !1, Ke = e.IN_PLACE || !1, Se = fs(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : ks, rt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : nt, ct = Us(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, st)), ut = Us(e, "HTML_INTEGRATION_POINTS", () => Q({}, lt));
		let t = Us(e, "CUSTOM_ELEMENT_HANDLING", () => Mo(null));
		if (we = Mo(null), ns(t, "tagNameCheck") && gt(t.tagNameCheck) && (we.tagNameCheck = t.tagNameCheck), ns(t, "attributeNameCheck") && gt(t.attributeNameCheck) && (we.attributeNameCheck = t.attributeNameCheck), ns(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (we.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), jo(we), Me && (ke = !1), Be && (ze = !0), Je && (k = Q({}, ys), j = Mo(null), Je.html === !0 && (Q(k, ps), Q(j, bs)), Je.svg === !0 && (Q(k, ms), Q(j, xs), Q(j, Cs)), Je.svgFilters === !0 && (Q(k, hs), Q(j, xs), Q(j, Cs)), Je.mathMl === !0 && (Q(k, _s), Q(j, Ss), Q(j, Cs))), De.tagCheck = null, De.attributeCheck = null, ns(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? De.tagCheck = e.ADD_TAGS : Ho(e.ADD_TAGS) && (k === A && (k = cs(k)), Q(k, e.ADD_TAGS, N))), ns(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? De.attributeCheck = e.ADD_ATTR : Ho(e.ADD_ATTR) && (j === Ce && (j = cs(j)), Q(j, e.ADD_ATTR, N))), ns(e, "ADD_FORBID_CONTENTS") && Ho(e.ADD_FORBID_CONTENTS) && (Ye === Xe && (Ye = cs(Ye)), Q(Ye, e.ADD_FORBID_CONTENTS, N)), Ge && (k["#text"] = !0), Pe && Q(k, [
			"html",
			"head",
			"body"
		]), k.table && (Q(k, ["tbody"]), delete Te.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw as("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw as("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = T;
			T = e.TRUSTED_TYPES_POLICY;
			try {
				ne = se("");
			} catch (e) {
				throw T = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (T = void 0, ne = "") : (T === void 0 && (T = le()), T && typeof ne == "string" && (ne = se("")));
		Ao && Ao(e), mt = e;
	}, vt = Q({}, [
		...ms,
		...hs,
		...gs
	]), yt = Q({}, [..._s, ...vs]), P = function(e, t, n) {
		return t.namespaceURI === nt ? e === "svg" : t.namespaceURI === et ? e === "svg" && (n === "annotation-xml" || ct[n]) : !!vt[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "math" : t.namespaceURI === tt ? e === "math" && ut[n] : !!yt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === tt && !ut[n] || t.namespaceURI === et && !ct[n] ? !1 : !yt[e] && (dt[e] || !vt[e]);
	}, St = function(e) {
		let t = te(e), n = Ko(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = Ko(w(r))) : (i = { namespaceURI: rt }, a = "template"), at[t] ? t === tt ? P(n, i, a) : t === et ? bt(n, i, a) : t === nt ? xt(n, i, a) : !!(ft === "application/xhtml+xml" && at[t]) : !1;
	}, Ct = function(e) {
		zo(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw as("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, wt = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, Tt = function(e) {
		kt(e);
		let t = g(e);
		t && Io(Bo(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && wt(e, r, i);
		}
	}, Et = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		zo(t.removed, {
			attribute: r || null,
			from: n
		});
		try {
			r ? m(n, r) : n.removeAttribute(e);
		} catch {
			try {
				n.removeAttribute(e);
			} catch {}
		}
		if (e === "is") {
			if (ze || Be) try {
				Ct(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Dt = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || j[N(i)] || wt(e, r, i);
		}
	}, Ot = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				Ht(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, kt = function(e) {
		Ot(e, (e) => {
			C(e) === $.element && Dt(e);
		});
	}, At = function(e) {
		let t = [];
		return Ot(e, (e) => {
			t.push(e);
		}, !0), t;
	}, jt = function(e, t) {
		let n = Uo();
		Io(At(e), (e) => {
			C(e) === $.element && Wo(n, e);
		}), Io(t, (e) => {
			C(e) === $.element && !Go(n, e) && Dt(e);
		});
	}, Mt = function(e, n) {
		let r = !1;
		return Io(t.removed, (t) => {
			t.element && (r ||= t.element === e, kt(t.element));
		}), jt(e, n), r;
	}, Nt = function(e, t) {
		let n = qe, r = null, i = () => (r || (r = Uo(), Io(t, (e) => {
			Wo(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = At(e), a = [e], o = !1;
			if (Io(r, (e) => {
				C(e) === $.documentFragment ? zo(a, e) : !o && !Go(n, e) && (o = Go(i(), e));
			}), !o) return;
			if (t >= 8) throw as("the tree kept changing during in-place sanitization; refusing to return it");
			Io(a, (t, n) => {
				Pt(t, n > 0, e, i());
			});
		}
	}, Pt = function(e, t, n, r) {
		let i = zt(e), a;
		for (; a = i.nextNode();) Go(r, a) && !Go(qe, a) && (t && Wt(E.uponSanitizeShadowNode, a, null), Xt(a, n), L(a, n), Ht(a.content) && rn(a.content));
	}, Ft = function(e) {
		qe && Wo(qe, e);
	}, It = function(e, t) {
		return Ne ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Lt = function(e) {
		Ne && Ot(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && is(Fs, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = N(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && It("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, Rt = function(e) {
		let t = null, r = null;
		if (Re) e = "<remove></remove>" + e;
		else {
			let t = Jo(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		ft === "application/xhtml+xml" && rt === nt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = T ? se(e) : e;
		if (rt === nt) try {
			t = new l().parseFromString(i, ft);
		} catch {}
		if (!t || !t.documentElement) {
			t = de.createDocument(rt, "template", null);
			try {
				t.documentElement.innerHTML = it ? ne : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), rt === nt ? me.call(t, Pe ? "html" : "body")[0] : Pe ? t.documentElement : a;
	}, zt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return fe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Bt = function(e) {
		return e = Yo(e, ge, " "), e = Yo(e, _e, " "), e = Yo(e, ve, " "), e;
	}, F = function(e) {
		Vt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && Io(t, (e) => {
			Ht(e.content) && F(e.content);
		});
	}, Vt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = fe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Bt(r.data), r = n.nextNode();
	}, I = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || N(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, Ht = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, Ut = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function Wt(e, n, r) {
		e.length !== 0 && Io(e, (e) => {
			e.call(t, n, r, mt);
		});
	}
	let Gt = function(e, t) {
		return !!(Ne && e.hasChildNodes() && !Ut(e.firstElementChild) && is(Ps, e.textContent) && is(Ps, e.innerHTML) || Ne && Rs[t] && C(e) === $.element && te(e) === nt && (Ut(e.firstElementChild) || typeof e.textContent == "string" && is(Ps, e.textContent)) || e.nodeType === $.processingInstruction || Ne && e.nodeType === $.comment && is(Fs, e.data));
	}, Kt = function(e, t) {
		return e instanceof RegExp ? is(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, qt = function(e, t, n) {
		if (!Te[t] && $t(t) && Kt(we.tagNameCheck, t)) return !1;
		if (Ge && !Ye[t] && !Ye[Ko(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return Ct(e), !0;
	}, Jt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? cs(t) : t;
	}, Yt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ke && kt(e), !0);
	}, Xt = function(e, n) {
		if (Ft(e), Wt(E.beforeSanitizeElements, e, null), Yt(e, n)) return !0;
		if (I(e)) return Ct(e), !0;
		let r = N(w(e));
		if (k = Jt(E.uponSanitizeElement, k, A, Ie), Wt(E.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: k
		}), Yt(e, n)) return !0;
		if (Gt(e, r)) return Ct(e), !0;
		if (Te[r] || !(De.tagCheck instanceof Function && De.tagCheck(r)) && !k[r]) {
			let t = qt(e, r, n);
			return t === !1 && (Wt(E.afterSanitizeElements, e, null), Yt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !St(e) || (r === "noscript" || r === "noembed" || r === "noframes") && is(Is, e.innerHTML)) return Ct(e), !0;
		if (Me && e.nodeType === $.text) {
			let n = Bt(e.textContent);
			e.textContent !== n && (zo(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Wt(E.afterSanitizeElements, e, null), Yt(e, n);
	}, Zt = function(e, t, r) {
		if (Ee[t] || It(t, e) || He && (t === "id" || t === "name") && (r in n || r in ht)) return !1;
		let i = j[t] || De.attributeCheck instanceof Function && De.attributeCheck(t, e);
		return ke && is(ye, t) || Oe && is(be, t) ? !0 : i ? M[t] || is(Se, Yo(r, xe, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Xo(r, "data:") === 0 && Ze[e] || Ae && !is(D, Yo(r, xe, "")) ? !0 : !r : $t(e) && Kt(we.tagNameCheck, e) && Kt(we.attributeNameCheck, t, e) || t === "is" && we.allowCustomizedBuiltInElements && Kt(we.tagNameCheck, r);
	}, Qt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), $t = function(e) {
		return !Qt[Ko(e)] && is(O, e);
	}, en = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return se(r);
			case "TrustedScriptURL": return ce(r);
		}
		return r;
	}, tn = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !I(e) || (Ct(e), !1);
		} catch {
			return Et(t, e), !1;
		}
	}, nn = function(e, t) {
		return Ne && is(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Ko(e) === "attributename" && Jo(t, "href") !== null;
	}, L = function(e, n) {
		if (Wt(E.beforeSanitizeAttributes, e, null), Yt(e, n)) return;
		let r = e.attributes;
		if (!r || I(e)) return;
		j = Jt(E.uponSanitizeAttribute, j, Ce, Le);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: j,
			forceKeepAttr: void 0
		}, a = r.length, o = N(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = N(s), d = l, f = s === "value" ? d : Zo(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Wt(E.uponSanitizeAttribute, e, i), f = i.attrValue, Ue && (u === "id" || u === "name") && Xo(f, We) !== 0 && (Et(s, e, n), f = We + f, p = !0), nn(s, f) ? Et(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !je && is(Ls, f) ? Et(s, e, n) : (Me && (f = Bt(f)), Zt(o, u, f) ? (f = en(o, u, c, f), f !== d && tn(e, s, c, f) && p && Ro(t.removed)) : Et(s, e, n)) : Et(s, e, n));
		}
		Wt(E.afterSanitizeAttributes, e, null), Yt(e, n);
	}, rn = function(e) {
		let t = null, n = zt(e);
		for (Wt(E.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Wt(E.uponSanitizeShadowNode, t, null), Xt(t, e), L(t, e), Ht(t.content) && rn(t.content), C(t) === $.element) {
			let e = v(t);
			Ht(e) && (R(e), rn(e));
		}
		Me && Vt(e), Wt(E.afterSanitizeShadowDOM, e, null);
	}, R = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				rn(e.shadow);
				continue;
			}
			let n = e.node, r = C(n) === $.element, i = g(n);
			if (i) for (let e = i.length - 1; e >= 0; --e) t.push({
				node: i[e],
				shadow: null
			});
			if (r) {
				let e = x ? x(n) : null;
				if (typeof e == "string" && N(e) === "template") {
					let e = n.content;
					Ht(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				Ht(e) && t.push({
					node: null,
					shadow: e
				}, {
					node: e,
					shadow: null
				});
			}
		}
	};
	return t.sanitize = function(e) {
		let n = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {}, i = null, a = null, o = null, s = null;
		if (it = !e, it && (e = "<!-->"), typeof e != "string" && !Ut(e) && (e = ls(e), typeof e != "string")) throw as("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Fe ? (k = Ie, j = Le) : _t(n), (E.uponSanitizeElement.length > 0 || E.uponSanitizeAttribute.length > 0) && (k = cs(k)), E.uponSanitizeAttribute.length > 0 && (j = cs(j)), t.removed = [];
		let c = Ke && typeof e != "string" && Ut(e), l = [], u = qe;
		if (c) {
			Lt(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = N(t);
				if (!k[n] || Te[n]) throw Tt(e), as("root node is forbidden and cannot be sanitized in-place");
			}
			if (I(e)) throw Tt(e), as("root node is clobbered and cannot be sanitized in-place");
			l = At(e), qe = Uo();
		} else if (Ut(e)) i = Rt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), R(i);
		else {
			if (!ze && !Me && !Pe && e.indexOf("<") === -1) return T && Ve ? se(e) : e;
			if (i = Rt(e), !i) return ze ? null : Ve ? ne : "";
		}
		i && Re && Ct(i.firstChild);
		let d = c ? e : i;
		try {
			c && R(e);
			let t = zt(d);
			for (; o = t.nextNode();) Xt(o, d), L(o, d), Ht(o.content) && rn(o.content);
			c && Nt(e, l);
		} catch (t) {
			throw c && (Tt(e), Mt(e, l)), t;
		} finally {
			qe = u;
		}
		if (c) {
			if (Mt(e, l)) throw as("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Me && F(e), e;
		}
		if (ze) {
			if (Me && F(i), Be) for (s = pe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (j.shadowroot || j.shadowrootmode) && (s = he.call(r, s, !0)), s;
		}
		let f = Pe ? i.outerHTML : i.innerHTML;
		return Pe && k["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && is(Ms, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Me && (f = Bt(f)), T && Ve ? se(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		_t(e), Fe = !0, Ie = k, Le = j;
	}, t.clearConfig = function() {
		mt = null, Fe = !1, Ie = null, Le = null, T = re, ne = "";
	}, t.isValidAttribute = function(e, t, n) {
		mt || _t({});
		let r = N(e), i = N(t);
		return typeof n == "string" && (nn(t, n) || !je && is(Ls, n)) ? !1 : Zt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && ns(E, e) && zo(E[e], t);
	}, t.removeHook = function(e, t) {
		if (ns(E, e)) {
			if (t !== void 0) {
				let n = Lo(E[e], t);
				return n === -1 ? void 0 : Vo(E[e], n, 1)[0];
			}
			return Ro(E[e]);
		}
	}, t.removeHooks = function(e) {
		ns(E, e) && (E[e] = []);
	}, t.removeAllHooks = function() {
		E = Vs();
	}, t;
}
var Gs = Ws();
// plugins/chat/format.js
function Ks(e) {
	let t = X.parse(e || "", {
		async: !1,
		gfm: !0,
		breaks: !1
	}), n = Gs.sanitize(t, {
		USE_PROFILES: { html: !0 },
		FORBID_TAGS: [
			"style",
			"form",
			"input",
			"button",
			"iframe",
			"object",
			"embed",
			"svg",
			"math",
			"img"
		],
		FORBID_ATTR: ["style"]
	}), r = new DOMParser().parseFromString(n, "text/html");
	for (let e of r.querySelectorAll("a")) {
		let t = !1;
		try {
			t = ["http:", "https:"].includes(new URL(e.href).protocol);
		} catch {}
		t || e.removeAttribute("href"), e.target = "_blank", e.rel = "noopener noreferrer";
	}
	return r.body.innerHTML;
}
function qs(e) {
	return e >= 1e6 ? (e / 1e6).toFixed(1) + "M" : e >= 1e3 ? Math.round(e / 1e3) + "k" : String(e);
}
var Js = (e) => "desk-chat-prompts." + e;
function Ys(e, t) {
	let n = [];
	try {
		let t = JSON.parse(localStorage.getItem(Js(e)) ?? "[]");
		Array.isArray(t) && (n = t.filter((e) => typeof e == "string"));
	} catch {}
	return [...new Set([...t, ...n.filter((e) => !t.includes(e))].filter(Boolean).reverse())].slice(0, 100);
}
function Xs(e, t) {
	try {
		let n = Ys(e, []).filter((e) => e !== t);
		localStorage.setItem(Js(e), JSON.stringify([...n.reverse(), t].slice(-100)));
	} catch {}
}
// plugins/chat/Block.svelte
var Zs = /* @__PURE__ */ W("<details class=\"prompt svelte-5n3hzh\"><summary class=\"svelte-5n3hzh\"> </summary><pre class=\"svelte-5n3hzh\"> </pre></details>"), Qs = /* @__PURE__ */ W("<div class=\"md svelte-5n3hzh\"></div>"), $s = /* @__PURE__ */ W("<div class=\"user svelte-5n3hzh\" data-testid=\"block-user\"><span class=\"tag svelte-5n3hzh\"> </span> <!></div>"), ec = /* @__PURE__ */ W("<div data-testid=\"block-assistant\"></div>"), tc = /* @__PURE__ */ W("<details class=\"reasoning svelte-5n3hzh\" data-testid=\"block-reasoning\"><summary class=\"svelte-5n3hzh\">Thinking</summary> <pre class=\"svelte-5n3hzh\"> </pre></details>"), nc = /* @__PURE__ */ W("<pre class=\"io svelte-5n3hzh\"> </pre>"), rc = /* @__PURE__ */ W("<pre class=\"io out svelte-5n3hzh\"> </pre>"), ic = /* @__PURE__ */ W("<!> <!>", 1), ac = /* @__PURE__ */ W("<div class=\"tool svelte-5n3hzh\" data-testid=\"block-tool\"><button class=\"tool-head svelte-5n3hzh\"><span class=\"tname svelte-5n3hzh\"> </span> <span class=\"tintent svelte-5n3hzh\"> </span> <span class=\"tstate svelte-5n3hzh\" data-testid=\"tool-state\"> </span></button> <!></div>"), oc = /* @__PURE__ */ W("<p class=\"error svelte-5n3hzh\" data-testid=\"block-error\"> </p>"), sc = /* @__PURE__ */ W("<pre class=\"svelte-5n3hzh\"> </pre>"), cc = /* @__PURE__ */ W("<details class=\"system svelte-5n3hzh\" data-testid=\"block-system\"><summary class=\"svelte-5n3hzh\"> </summary> <!></details>");
function lc(e, t) {
	We(t, !0);
	let n = /* @__PURE__ */ F(!1);
	function r(e, t = 140) {
		let n = (e || "").replace(/\s+/g, " ").trim();
		return n.length > t ? n.slice(0, t) + "…" : n;
	}
	function i(e) {
		try {
			return JSON.stringify(JSON.parse(e), null, 2);
		} catch {
			return e;
		}
	}
	function a(e) {
		let t = /"intent"\s*:\s*"((?:[^"\\]|\\.)*)"/.exec(e || "");
		return t ? t[1] : "";
	}
	let o = {
		operator: "operator",
		desk: "Desk user",
		controller: "controller prompt",
		message: "message"
	};
	var s = wr(), c = rn(s), l = (e) => {
		var n = $s(), i = L(n), a = R(i, !0), s = z(i, 2), c = (e) => {
			var n = Zs(), i = L(n), a = R(i, !0), o = R(z(i), !0);
			A(n), B((e) => {
				K(a, e), K(o, t.block.text);
			}, [() => r(t.block.text, 120)]), G(e, n);
		}, l = (e) => {
			var n = Qs();
			qr(n, () => Ks(t.block.text), !0), A(n), G(e, n);
		};
		q(s, (e) => {
			t.block.source === "controller" ? e(c) : e(l, -1);
		}), A(n), B(() => {
			ti(n, "data-source", t.block.source), K(a, o[t.block.source] || t.block.source);
		}), G(e, n);
	}, u = (e) => {
		var n = ec();
		let r;
		qr(n, () => Ks(t.block.text), !0), A(n), B(() => r = Xr(n, 1, "assistant md svelte-5n3hzh", null, r, { live: t.block.live })), G(e, n);
	}, d = (e) => {
		var n = tc(), r = R(z(L(n), 2), !0);
		A(n), B(() => {
			n.open = t.block.live, K(r, t.block.text);
		}), G(e, n);
	}, f = (e) => {
		var o = ac(), s = L(o), c = L(s), l = R(c, !0), u = z(c, 2), d = R(u, !0), f = R(z(u, 2), !0);
		A(s);
		var p = z(s, 2), m = (e) => {
			var n = ic(), r = rn(n), a = (e) => {
				var n = nc(), r = R(n, !0);
				B((e) => K(r, e), [() => i(t.block.input)]), G(e, n);
			};
			q(r, (e) => {
				t.block.input && e(a);
			});
			var o = z(r, 2), s = (e) => {
				var n = rc(), r = R(n, !0);
				B(() => K(r, t.block.output)), G(e, n);
			};
			q(o, (e) => {
				t.block.output && e(s);
			}), G(e, n);
		};
		q(p, (e) => {
			U(n) && e(m);
		}), A(o), B((e) => {
			ti(o, "data-state", t.block.state), ti(o, "data-tool", t.block.id), ti(s, "aria-expanded", U(n)), K(l, t.block.name), K(d, e), K(f, t.block.state);
		}, [() => a(t.block.input) || r(t.block.input, 80)]), mr("click", s, () => I(n, !U(n))), G(e, o);
	}, p = (e) => {
		var n = oc(), r = R(n, !0);
		B(() => K(r, t.block.text)), G(e, n);
	}, m = (e) => {
		var n = cc(), i = L(n), a = R(i, !0), o = z(i, 2), s = (e) => {
			var n = sc(), r = R(n, !0);
			B(() => K(r, t.block.text)), G(e, n);
		}, c = /* @__PURE__ */ pt(() => t.block.text.length > 160 || t.block.text.includes("\n"));
		q(o, (e) => {
			U(c) && e(s);
		}), A(n), B((e) => K(a, e), [() => r(t.block.text, 160)]), G(e, n);
	};
	q(c, (e) => {
		t.block.kind === "user" ? e(l) : t.block.kind === "assistant" ? e(u, 1) : t.block.kind === "reasoning" ? e(d, 2) : t.block.kind === "tool" ? e(f, 3) : t.block.kind === "error" ? e(p, 4) : e(m, -1);
	}), G(e, s), Ge();
}
hr(["click"]);
// plugins/chat/Pane.svelte
var uc = /* @__PURE__ */ W("<span> </span>"), dc = /* @__PURE__ */ W("· <a class=\"mono svelte-wy7o3a\"> </a>", 1), fc = /* @__PURE__ */ W("<button class=\"older svelte-wy7o3a\" data-testid=\"load-older\"> </button>"), pc = /* @__PURE__ */ W("<p class=\"err svelte-wy7o3a\"> </p>"), mc = /* @__PURE__ */ W("<p class=\"none svelte-wy7o3a\">No messages yet.</p>"), hc = /* @__PURE__ */ W("<p class=\"status svelte-wy7o3a\" data-testid=\"chat-working\"><span class=\"spin svelte-wy7o3a\"></span> </p>"), gc = /* @__PURE__ */ W("<label class=\"toggle svelte-wy7o3a\" data-testid=\"answer-toggle\"><input type=\"checkbox\"/> answer the question</label>"), _c = /* @__PURE__ */ W("<button type=\"button\" data-testid=\"resolve\" class=\"svelte-wy7o3a\"> </button>"), vc = /* @__PURE__ */ W("<div class=\"item svelte-wy7o3a\" data-testid=\"chat-item\"><div class=\"itxt svelte-wy7o3a\"><span class=\"pill svelte-wy7o3a\"> </span> <span class=\"ititle\"> </span></div> <div class=\"iacts svelte-wy7o3a\"><!> <!></div></div>"), yc = /* @__PURE__ */ W("<span class=\"model svelte-wy7o3a\"> </span>"), bc = /* @__PURE__ */ W("<span class=\"usage svelte-wy7o3a\"> </span>"), xc = /* @__PURE__ */ W("<span class=\"notice svelte-wy7o3a\" data-testid=\"chat-notice\"> </span>"), Sc = /* @__PURE__ */ W("<button type=\"button\" class=\"interrupt svelte-wy7o3a\" title=\"Cancel the turn and send now (Ctrl+Enter)\" data-testid=\"chat-interrupt\">Interrupt</button>"), Cc = /* @__PURE__ */ W("<button type=\"button\" class=\"stop svelte-wy7o3a\" title=\"Cancel the turn (Esc)\" data-testid=\"chat-stop\">Stop</button> <!>", 1), wc = /* @__PURE__ */ W("<form class=\"composer svelte-wy7o3a\"><textarea rows=\"2\" aria-label=\"Message\" enterkeyhint=\"send\" data-testid=\"chat-input\" class=\"svelte-wy7o3a\"></textarea> <div class=\"actions svelte-wy7o3a\"><button type=\"button\" class=\"history svelte-wy7o3a\" aria-label=\"Previous sent message\" title=\"Previous sent message (Up arrow)\">↑</button> <button type=\"button\" class=\"history svelte-wy7o3a\" aria-label=\"Next sent message\" title=\"Next sent message (Down arrow)\">↓</button> <span class=\"metadata svelte-wy7o3a\" data-testid=\"chat-usage\"><!> <!></span> <!> <span class=\"grow svelte-wy7o3a\"></span> <!> <button type=\"submit\" class=\"send svelte-wy7o3a\" data-testid=\"chat-send\"> </button></div></form>"), Tc = /* @__PURE__ */ W("<p class=\"ro svelte-wy7o3a\" data-testid=\"chat-readonly\"> </p>"), Ec = /* @__PURE__ */ W("<section class=\"pane svelte-wy7o3a\" data-testid=\"chat-pane\"><header class=\"head svelte-wy7o3a\"><button class=\"x close svelte-wy7o3a\" aria-label=\"Close chat\" title=\"Close\" data-testid=\"session-panel-close\">✕</button> <div class=\"who svelte-wy7o3a\"><div class=\"name svelte-wy7o3a\"> </div> <div class=\"meta svelte-wy7o3a\"><span class=\"mono svelte-wy7o3a\"> </span> <!> <!></div></div> <span class=\"state svelte-wy7o3a\" data-testid=\"chat-status\"> </span></header> <div class=\"scroll svelte-wy7o3a\" data-testid=\"journal\"><!> <!> <!> <!></div> <!> <!></section>");
function Dc(e, t) {
	We(t, !0);
	let n = pi(t, "context", 19, () => ({})), r = /* @__PURE__ */ pt(() => String(n().session || "")), i = /* @__PURE__ */ pt(() => String(n().item || "")), a = /* @__PURE__ */ pt(() => U(i).split(":")[0] || ""), o = /* @__PURE__ */ pt(() => U(i).slice(U(a).length + 1)), s = /* @__PURE__ */ F(null);
	gn(() => {
		if (!U(r)) return;
		let e = new vi(t.api, U(r));
		return I(s, e, !0), e.open(), () => e.close();
	});
	let c = /* @__PURE__ */ F(null), l = /* @__PURE__ */ F(""), u = /* @__PURE__ */ F(Jt({
		write: !1,
		reason: ""
	}));
	gn(() => {
		if (!U(r)) return;
		let e = !1;
		return I(c, null), I(l, ""), I(u, {
			write: !1,
			reason: ""
		}, !0), t.api.get("api/sessions?since=720h").then((t) => {
			e || I(c, t.find((e) => e.id === U(r)) || null, !0);
		}).catch(() => {}), t.api.get("api/chat").then((t) => {
			if (e) return;
			let n = (t.sessions || []).find((e) => e.id === U(r));
			I(l, n?.alive ? n.key : "", !0);
		}).catch(() => {}), t.api.session.access(U(r)).then((t) => {
			e || I(u, {
				write: !!t.write,
				reason: t.reason || ""
			}, !0);
		}).catch(() => {}), () => {
			e = !0;
		};
	});
	let d = /* @__PURE__ */ F(null);
	gn(() => {
		if (I(d, null), !U(o) || [
			"pool",
			"deploy",
			"rollback",
			"config",
			"config-repo",
			"ingest",
			"resource"
		].includes(U(a))) return;
		let e = !1;
		return t.api.get("api/ticket/" + encodeURIComponent(U(o))).then((t) => {
			e || I(d, t.ticket, !0);
		}).catch(() => {}), () => {
			e = !0;
		};
	});
	let f = /* @__PURE__ */ pt(() => !!U(s)?.busy), p = /* @__PURE__ */ pt(() => {
		let e = U(s)?.blocks ?? [];
		for (let t = e.length - 1; t >= 0; t--) {
			let n = e[t];
			if (n.kind === "tool" && (n.state === "running" || n.state === "input")) return n.name;
		}
		return "";
	}), m = /* @__PURE__ */ pt(() => U(u).write || !!U(l)), h = /* @__PURE__ */ pt(() => U(l).startsWith("escalation:") ? "Escalation " + U(l).slice(11) : U(l).startsWith("attention:") ? "Resolve " + U(i) : U(l).startsWith("desk:") ? "Desk chat" : U(c)?.ticket?.title || "Session"), g = /* @__PURE__ */ F(!0), _ = /* @__PURE__ */ pt(() => U(a) === "question" && U(g)), v = /* @__PURE__ */ F(""), y = /* @__PURE__ */ F(!1), b = /* @__PURE__ */ F(""), x = /* @__PURE__ */ F(void 0), S = /* @__PURE__ */ F(!0), C = /* @__PURE__ */ F(void 0), w = /* @__PURE__ */ F(-1), ee = [], te = "", T = null;
	function ne(e) {
		if (U(w) < 0) {
			if (!e) return;
			ee = Ys(U(r), U(s)?.prompts() ?? []), te = U(v);
		}
		let t = e ? Math.min(U(w) + 1, ee.length - 1) : U(w) - 1;
		t < -1 || !ee.length || (I(w, t, !0), I(v, t < 0 ? te : ee[t], !0), ar().then(() => {
			U(C)?.focus(), U(C)?.setSelectionRange(U(v).length, U(v).length), ae();
		}));
	}
	async function re() {
		U(s) && U(x) && (I(S, !1), T = {
			height: U(x).scrollHeight,
			top: U(x).scrollTop
		}, await U(s).loadOlder(), await ar(), T && U(x) && (U(x).scrollTop = T.top + U(x).scrollHeight - T.height, T = null));
	}
	function ie() {
		U(x) && I(S, U(x).scrollHeight - U(x).scrollTop - U(x).clientHeight < 80);
	}
	gn(() => {
		U(s)?.version, U(f), U(S) && !T && ar().then(() => {
			U(x) && (U(x).scrollTop = U(x).scrollHeight);
		});
	});
	function ae() {
		U(C) && (U(C).style.height = "auto", U(C).style.height = Math.min(U(C).scrollHeight, Math.round(window.innerHeight * .4)) + "px");
	}
	async function oe(e) {
		let n = U(v).trim();
		if (n && !U(y)) {
			I(y, !0), I(b, "");
			try {
				U(_) ? (await t.api.call("question.answer", {
					id: U(o),
					text: n
				}), I(b, "Answer sent to the asking session")) : (await t.api.session.send(U(l) ? {
					key: U(l),
					text: n,
					interrupt: e
				} : {
					session: U(r),
					text: n,
					interrupt: e
				}), U(f) && !e && I(b, "Queued for after this turn")), Xs(U(r), n), I(w, -1), I(v, ""), I(S, !0), await ar(), ae();
			} catch (e) {
				I(b, String(e.message || e), !0);
			} finally {
				I(y, !1);
			}
		}
	}
	async function se() {
		I(b, "");
		try {
			(await t.api.session.cancel(U(r))).cancelled || I(b, "No turn running");
		} catch (e) {
			I(b, String(e.message || e), !0);
		}
	}
	function ce(e) {
		if (e.isComposing) return;
		let t = e.currentTarget;
		e.key === "ArrowUp" && !e.shiftKey && t.selectionStart === t.selectionEnd && (U(w) >= 0 || !U(v).slice(0, t.selectionStart).includes("\n")) ? (e.preventDefault(), ne(!0)) : e.key === "ArrowDown" && !e.shiftKey && U(w) >= 0 ? (e.preventDefault(), ne(!1)) : e.key === "Enter" && !e.shiftKey ? (e.preventDefault(), oe(e.ctrlKey || e.metaKey)) : e.key === "Escape" && U(f) && (e.preventDefault(), e.stopPropagation(), se());
	}
	let le = /* @__PURE__ */ pt(() => {
		if (!U(a)) return [];
		let e = (e, n) => () => t.api.call(e, n), n = () => t.api.post("api/attention/ack", {
			kind: U(a),
			id: U(o)
		});
		switch (U(a)) {
			case "run": return [{
				label: "Resume run",
				run: e("run.resume", { id: U(o) }),
				ask: "Resume run " + U(o) + "?"
			}];
			case "pool": return [{
				label: "Unquarantine pool",
				run: e("pool.unquarantine", { pool: U(o) }),
				ask: "Unquarantine pool " + U(o) + "?"
			}];
			case "wait": return [{
				label: "Clear wait",
				run: e("wait.clear", {
					id: U(o),
					reason: "cleared in Desk"
				}),
				ask: "Clear the wait on " + U(o) + "?"
			}];
			case "decision": return [{
				label: "Approve as target",
				run: async () => (await t.api.call("label.add", {
					id: U(o),
					label: "target"
				}), t.api.call("label.remove", {
					id: U(o),
					label: "proposed"
				})),
				ask: "Approve " + U(o) + ": relabel proposed to target?"
			}, {
				label: "Close proposal",
				run: e("ticket.close", {
					id: U(o),
					resolution: "rejected"
				}),
				ask: "Close proposal " + U(o) + "?"
			}];
			case "ticket":
			case "question":
			case "stale": return [{
				label: "Close ticket",
				run: e("ticket.close", { id: U(o) }),
				ask: "Close " + U(o) + " as done?"
			}];
			default: return [{
				label: "Acknowledge",
				run: n,
				ask: "Acknowledge " + U(i) + "? It is hidden until it recurs."
			}];
		}
	}), ue = /* @__PURE__ */ F(!1);
	async function de(e) {
		if (await t.api.confirm(e.ask)) {
			I(ue, !0);
			try {
				await e.run(), t.api.toast(e.label + ": done");
			} catch (n) {
				t.api.toast(e.label + " failed: " + String(n.message || n));
			} finally {
				I(ue, !1);
			}
		}
	}
	var fe = Ec(), pe = L(fe), me = L(pe), he = z(me, 2), E = L(he), ge = R(E, !0), _e = z(E, 2), ve = L(_e), ye = R(ve, !0), be = z(ve, 2), D = (e) => {
		var t = uc(), n = R(t);
		B(() => K(n, `· ${U(c).ticket.pool ?? ""}`)), G(e, t);
	};
	q(be, (e) => {
		U(c)?.ticket?.pool && e(D);
	});
	var xe = z(be, 2), O = (e) => {
		var n = dc(), r = z(rn(n)), i = R(r, !0);
		B(() => {
			ti(r, "href", "?ticket=" + U(c).current), K(i, U(c).current);
		}), mr("click", r, (e) => {
			e.preventDefault(), t.api.openTicket(U(c).current);
		}), G(e, n);
	};
	q(xe, (e) => {
		U(c)?.current && e(O);
	}), A(_e), A(he);
	var Se = R(z(he, 2), !0);
	A(pe);
	var k = z(pe, 2), Ce = L(k), we = (e) => {
		var t = fc(), n = R(t, !0);
		B(() => {
			t.disabled = U(s).loadingOlder, K(n, U(s).loadingOlder ? "Loading history…" : `Load older messages (${U(s).earlier} earlier)`);
		}), mr("click", t, re), G(e, t);
	};
	q(Ce, (e) => {
		U(s) && U(s).earlier > 0 && e(we);
	});
	var Te = z(Ce, 2), Ee = (e) => {
		var t = pc(), n = R(t, !0);
		B(() => K(n, U(s).error)), G(e, t);
	};
	q(Te, (e) => {
		U(s)?.error && e(Ee);
	});
	var De = z(Te, 2);
	Vr(De, 17, () => U(s)?.blocks ?? [], Lr, (e, t) => {
		lc(e, { get block() {
			return U(t);
		} });
	}, (e) => {
		G(e, mc());
	});
	var Oe = z(De, 2), ke = (e) => {
		var t = hc(), n = z(L(t), 1, !0);
		A(t), B(() => K(n, U(p) || "Working")), G(e, t);
	};
	q(Oe, (e) => {
		U(f) && e(ke);
	}), A(k), ui(k, (e) => I(x, e), () => U(x));
	var Ae = z(k, 2), je = (e) => {
		var t = vc(), n = L(t), r = L(n), i = R(r, !0), s = R(z(r, 2), !0);
		A(n);
		var c = z(n, 2), l = L(c), u = (e) => {
			var t = gc(), n = L(t);
			ei(n), j(), A(t), oi(n, () => U(g), (e) => I(g, e)), G(e, t);
		};
		q(l, (e) => {
			U(a) === "question" && e(u);
		}), Vr(z(l, 2), 17, () => U(le), (e) => e.label, (e, t) => {
			var n = _c(), r = R(n, !0);
			B(() => {
				n.disabled = U(ue), K(r, U(t).label);
			}), mr("click", n, () => de(U(t))), G(e, n);
		}), A(c), A(t), B(() => {
			K(i, U(a)), K(s, U(d)?.title || U(o));
		}), G(e, t);
	};
	q(Ae, (e) => {
		U(i) && e(je);
	});
	var Me = z(Ae, 2), Ne = (e) => {
		var t = wc(), n = L(t);
		tt(n), ui(n, (e) => I(C, e), () => U(C));
		var r = z(n, 2), i = L(r), a = z(i, 2), o = z(a, 2), l = L(o), u = (e) => {
			var t = yc(), n = R(t, !0);
			B(() => K(n, U(c).model)), G(e, t);
		};
		q(l, (e) => {
			U(c)?.model && e(u);
		});
		var d = z(l, 2), p = (e) => {
			var t = bc(), n = R(t);
			B((e, t) => K(n, `${e ?? ""} / ${t ?? ""} tokens`), [() => qs(U(s).usage.used), () => qs(U(s).usage.size)]), G(e, t);
		};
		q(d, (e) => {
			U(s)?.usage && e(p);
		}), A(o);
		var m = z(o, 2), h = (e) => {
			var t = xc(), n = R(t, !0);
			B(() => K(n, U(b))), G(e, t);
		};
		q(m, (e) => {
			U(b) && e(h);
		});
		var g = z(m, 4), x = (e) => {
			var t = Cc(), n = rn(t), r = z(n, 2), i = (e) => {
				var t = Sc();
				B((e) => t.disabled = e, [() => !U(v).trim() || U(y)]), mr("click", t, () => oe(!0)), G(e, t);
			};
			q(r, (e) => {
				U(_) || e(i);
			}), mr("click", n, se), G(e, t);
		};
		q(g, (e) => {
			U(f) && e(x);
		});
		var S = z(g, 2), ee = R(S, !0);
		A(r), A(t), B((e) => {
			ti(n, "placeholder", U(_) ? "Answer the question" : U(f) ? "Queued for after this turn" : "Message"), a.disabled = U(w) < 0, S.disabled = e, K(ee, U(_) ? "Answer" : U(f) ? "Queue" : "Send");
		}, [() => !U(v).trim() || U(y)]), pr("submit", t, (e) => {
			e.preventDefault(), oe(!1);
		}), mr("input", n, () => {
			I(w, -1), ae();
		}), mr("keydown", n, ce), ai(n, () => U(v), (e) => I(v, e)), mr("click", i, () => ne(!0)), mr("click", a, () => ne(!1)), G(e, t);
	}, Pe = (e) => {
		var t = Tc(), n = R(t);
		B(() => K(n, `Read-only: ${(U(u).reason || "this session does not accept input") ?? ""}`)), G(e, t);
	};
	q(Me, (e) => {
		U(m) || U(_) ? e(Ne) : e(Pe, -1);
	}), A(fe), B(() => {
		ti(fe, "aria-label", U(h)), K(ge, U(h)), K(ye, U(r)), K(Se, U(s)?.status ?? "connecting");
	}), mr("click", me, () => t.api.closeSession()), pr("scroll", k, ie), G(e, fe), Ge();
}
hr([
	"click",
	"input",
	"keydown"
]);
// plugins/chat/ChatList.svelte
var Oc = /* @__PURE__ */ W("<div class=\"err svelte-162crgi\"> </div>"), kc = /* @__PURE__ */ W("<a><span></span> <span class=\"name svelte-162crgi\"> </span> <span class=\"mono faint id svelte-162crgi\"> </span></a>"), Ac = /* @__PURE__ */ W("<div class=\"faint pad0 svelte-162crgi\">No open operator sessions.</div>"), jc = /* @__PURE__ */ W("<div class=\"card empty\" data-testid=\"escalations-empty\"><div class=\"muted\">No escalation sessions.</div></div>"), Mc = /* @__PURE__ */ W("<tr data-testid=\"escalation-row\"><td class=\"c-title\"> <div class=\"faint mono\"> <!></div></td><td class=\"c-st\"><span> </span></td><td class=\"muted full\"> </td><td class=\"c-st\"><span> </span></td><td class=\"muted c-time\"> </td></tr>"), Nc = /* @__PURE__ */ W("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"escalations\"><thead><tr><th>Ticket</th><th>Cause</th><th>Why</th><th>State</th><th>Age</th></tr></thead><tbody></tbody></table></div>"), Pc = /* @__PURE__ */ W("<header class=\"vh svelte-162crgi\"><h1>Chat</h1> <span class=\"sub svelte-162crgi\">pool <span class=\"mono\"> </span></span></header> <div class=\"list svelte-162crgi\"><form class=\"new svelte-162crgi\"><textarea rows=\"3\" placeholder=\"What should the operator look at?\" data-testid=\"chat-new-input\" class=\"svelte-162crgi\"></textarea> <div class=\"row svelte-162crgi\"><button type=\"submit\" class=\"primary\" data-testid=\"chat-new\">New session</button></div></form> <!> <div data-testid=\"chat-list\"></div> <h2 class=\"st\">Escalations</h2> <!></div>", 1);
function Fc(e, t) {
	We(t, !0);
	let n = /* @__PURE__ */ F(null), r = /* @__PURE__ */ F(null), i = /* @__PURE__ */ F(""), a = /* @__PURE__ */ F(""), o = /* @__PURE__ */ F(!1), s = /* @__PURE__ */ F("");
	async function c() {
		try {
			await (async (e) => {
				var t = y(e, 2);
				I(n, t[0], !0), I(r, t[1], !0);
			})(await Promise.all([t.api.get("api/chat"), t.api.get("api/escalations")])), I(i, "");
		} catch (e) {
			I(i, String(e.message || e), !0);
		}
	}
	gn(() => {
		I(s, t.api.params().session || "", !0), c();
		let e = 0, n = t.api.subscribe(() => {
			I(s, t.api.params().session || "", !0), e ||= window.setTimeout(() => {
				e = 0, c();
			}, 1e3);
		});
		return () => {
			n(), clearTimeout(e);
		};
	});
	let l = /* @__PURE__ */ pt(() => (U(n)?.sessions ?? []).filter((e) => !e.key.startsWith("escalation:"))), u = /* @__PURE__ */ pt(() => (U(r) ?? []).filter((e) => e.open)), d = /* @__PURE__ */ pt(() => (U(r) ?? []).filter((e) => !e.open));
	function f(e) {
		return e.startsWith("attention:") ? "Resolve " + e.slice(10) : e.startsWith("desk:") ? "Desk chat" : e.split(":")[0] || "Session";
	}
	let p = {
		run: "run failed",
		pool: "pool quarantined",
		wait: "wait timed out",
		guard: "needs approval",
		label: "needs human"
	};
	function m(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	function h(e, n) {
		I(s, e, !0), t.api.openSession(e, n ? { item: n } : void 0);
	}
	async function g() {
		if (U(a).trim()) {
			I(o, !0), I(i, "");
			try {
				let e = await t.api.post("api/chat/new", { text: U(a) });
				I(a, ""), h(e.session);
			} catch (e) {
				I(i, String(e.message || e), !0);
			} finally {
				I(o, !1);
			}
		}
	}
	var _ = Pc(), v = rn(_), b = z(L(v), 2), x = R(z(L(b)), !0);
	A(b), A(v);
	var S = z(v, 2), C = L(S), w = L(C);
	tt(w);
	var ee = R(z(w, 2));
	A(C);
	var te = z(C, 2), T = (e) => {
		var t = Oc(), n = R(t, !0);
		B(() => K(n, U(i))), G(e, t);
	};
	q(te, (e) => {
		U(i) && e(T);
	});
	var ne = z(te, 2);
	Vr(ne, 21, () => U(l), (e) => e.id, (e, t) => {
		var n = kc();
		let r;
		var i = L(n);
		let a;
		var o = z(i, 2), c = R(o, !0), l = R(z(o, 2), !0);
		A(n), B((e) => {
			ti(n, "href", "?view=chat&session=" + U(t).id), r = Xr(n, 1, "item svelte-162crgi", null, r, { active: U(s) === U(t).id }), a = Xr(i, 1, "dot", null, a, { accent: U(t).alive }), K(c, e), K(l, U(t).id);
		}, [() => f(U(t).key)]), mr("click", n, (e) => {
			e.preventDefault(), h(U(t).id, U(t).key.startsWith("attention:") ? U(t).key.slice(10) : "");
		}), G(e, n);
	}, (e) => {
		G(e, Ac());
	}), A(ne);
	var re = z(ne, 4), ie = (e) => {
		G(e, jc());
	}, ae = (e) => {
		var t = Nc(), n = L(t), r = z(L(n));
		Vr(r, 21, () => [...U(u), ...U(d)], (e) => e.ticket, (e, t) => {
			var n = Mc();
			let r;
			var i = L(n), a = L(i, !0), o = z(a), c = L(o, !0), l = z(c), u = (e) => {
				var n = Cr();
				B(() => K(n, `· ${U(t).held ?? ""}`)), G(e, n);
			};
			q(l, (e) => {
				U(t).held && U(t).held !== U(t).ticket && e(u);
			}), A(o), A(i);
			var d = z(i), f = L(d), g = R(f, !0);
			A(d);
			var _ = z(d), v = R(_, !0), y = z(_), b = L(y), x = R(b, !0);
			A(y);
			var S = R(z(y), !0);
			A(n), B((e) => {
				r = Xr(n, 1, "click", null, r, { sel: U(s) === U(t).session }), K(a, U(t).title), K(c, U(t).ticket), Xr(f, 1, `pill ${U(t).open ? "warn" : ""}`), K(g, p[U(t).cause] ?? U(t).cause ?? "escalation"), K(v, U(t).why), Xr(b, 1, `pill ${U(t).open ? "accent" : "ok"}`), K(x, U(t).open ? "open" : "closed" + (U(t).resolution ? " · " + U(t).resolution : "")), K(S, e);
			}, [() => m(U(t).open ? U(t).since : U(t).closed)]), mr("click", n, () => h(U(t).session, "ticket:" + U(t).ticket)), G(e, n);
		}), A(r), A(n), A(t), G(e, t);
	};
	q(re, (e) => {
		U(r) && U(u).length + U(d).length === 0 ? e(ie) : U(r) && e(ae, 1);
	}), A(S), B((e) => {
		K(x, U(n)?.pool ?? "…"), ee.disabled = e;
	}, [() => U(o) || !U(a).trim()]), pr("submit", C, (e) => {
		e.preventDefault(), g();
	}), ai(w, () => U(a), (e) => I(a, e)), G(e, _), Ge();
}
hr(["click"]);
// plugins/chat/DrawerChat.svelte
var Ic = /* @__PURE__ */ W("<div class=\"wrap svelte-1i9xi0d\" data-testid=\"escalation-chat\"><div class=\"bar svelte-1i9xi0d\"><a class=\"svelte-1i9xi0d\">open in side panel</a></div> <div class=\"frame svelte-1i9xi0d\"><!></div></div>");
function Lc(e, t) {
	We(t, !0);
	let n = pi(t, "context", 19, () => ({})), r = /* @__PURE__ */ pt(() => n().ticket || {}), i = (e) => typeof e?.meta == "string" ? JSON.parse(e.meta) : e?.meta || {}, a = /* @__PURE__ */ pt(() => {
		if (U(r).kind === "session") return U(r).id;
		let e = i(U(r)).escalation || i(U(r)).gateway;
		return e?.session ? e.session : U(r).holder && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(U(r).holder) ? U(r).holder : "";
	});
	var o = wr(), s = rn(o), c = (e) => {
		var n = Ic(), r = L(n), i = R(r), o = z(r, 2), s = L(o);
		{
			let e = /* @__PURE__ */ pt(() => ({ session: U(a) }));
			Dc(s, {
				get api() {
					return t.api;
				},
				get context() {
					return U(e);
				}
			});
		}
		A(o), A(n), B(() => ti(i, "href", "?session=" + U(a))), mr("click", i, (e) => {
			e.preventDefault(), t.api.openSession(U(a));
		}), G(e, n);
	};
	q(s, (e) => {
		U(a) && e(c);
	}), G(e, o), Ge();
}
hr(["click"]);
// plugins/chat/index.svelte.js
var Rc = new URL("./style.css", import.meta.url).href;
function zc(e) {
	return class extends HTMLElement {
		#e = /* @__PURE__ */ F(Jt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return U(this.#e);
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
			this.props.context = e;
		}
		get context() {
			return this.props.context;
		}
		set config(e) {
			this.props.config = e;
		}
		get config() {
			return this.props.config;
		}
		connectedCallback() {
			if (!document.querySelector(`link[href="${Rc}"]`)) {
				let e = document.createElement("link");
				e.rel = "stylesheet", e.href = Rc, document.head.append(e);
			}
			this.style.display = "contents", this.app = jr(e, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && Fr(this.app), this.app = null;
		}
	};
}
for (let [e, t] of [
	["desk-chat-pane", Dc],
	["desk-chat-view", Fc],
	["desk-chat-drawer", Lc]
]) customElements.get(e) || customElements.define(e, zc(t));
