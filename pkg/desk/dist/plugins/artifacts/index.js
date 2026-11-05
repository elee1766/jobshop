// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = "http://www.w3.org/2000/svg", i = "http://www.w3.org/1998/Math/MathML", a = Array.isArray, o = Array.prototype.indexOf, s = Array.prototype.includes, c = Array.from, l = Object.defineProperty, u = Object.getOwnPropertyDescriptor, d = Object.getOwnPropertyDescriptors, f = Object.prototype, p = Array.prototype, m = Object.getPrototypeOf, h = Object.isExtensible, g = () => {};
function _(e) {
	return typeof e?.then == "function";
}
function v(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function y() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var b = 1024, x = 2048, S = 4096, C = 8192, w = 16384, ee = 32768, te = 1 << 25, T = 65536, ne = 1 << 19, re = 1 << 20, ie = 1 << 25, ae = 1 << 21, oe = 1 << 22, se = 1 << 23, ce = Symbol("$state"), le = Symbol("component"), ue = Symbol(""), de = Symbol("attributes"), fe = Symbol("class"), pe = Symbol("style"), me = Symbol("text"), he = Symbol("form reset"), E = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), ge = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function _e() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ve(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ye() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var D = !1;
function be(e) {
	D = e;
}
var O;
function k(t) {
	if (t === null) throw ve(), e;
	return O = t;
}
function xe() {
	return k(/* @__PURE__ */ Xt(O));
}
function A(t) {
	if (D) {
		if (/* @__PURE__ */ Xt(O) !== null) throw ve(), e;
		O = t;
	}
}
function Se(e = 1) {
	if (D) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ Xt(n);
		O = n;
	}
}
function j(e = !0) {
	for (var t = 0, n = O;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Xt(n);
		e && n.remove(), n = i;
	}
}
function Ce(t) {
	if (!t || t.nodeType !== 8) throw ve(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function we(e) {
	return e === this.v;
}
function Te(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Ee(e) {
	return !Te(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function De() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Oe(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function ke(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Ae() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function je(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Me() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Ne() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Pe() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Ie() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var M = null;
function Le(e) {
	M = e;
}
function Re(e, t = !1, n) {
	M = {
		p: M,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: V,
		l: null
	};
}
function ze(e) {
	var t = M, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) fn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, M = t.p, Be(e);
}
function Be(e = {}) {
	return l(e, le, { value: !0 }), e;
}
function Ve() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var He = [];
function Ue() {
	var e = He;
	He = [], v(e);
}
function We(e) {
	if (He.length === 0 && !gt) {
		var t = He;
		queueMicrotask(() => {
			t === He && Ue();
		});
	}
	He.push(e);
}
function Ge() {
	for (; He.length > 0;) Ue();
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ke = ~(x | S | b);
function N(e, t) {
	e.f = e.f & Ke | t;
}
function qe(e) {
	e.f & 512 || e.deps === null ? N(e, b) : N(e, S);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Je(e, t) {
	e.f & 2048 ? t.set(e, x) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, S), N(e, b);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
var Ye = !1;
function Xe() {
	Ye || (Ye = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[he]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Ze(e) {
	var t = B, n = V;
	Pn(null), Fn(null);
	try {
		return e();
	} finally {
		Pn(t), Fn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Qe(e, t, n, r) {
	let i = Ve() ? nt : ot;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = V, c = $e(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				an(e, s);
			}
			et();
		}
	}
	var d = tt();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ it(e))).then(u).catch((e) => an(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), et());
	}) : f();
}
function $e() {
	var e = V, t = B, n = M, r = P;
	return function(i = !0) {
		Fn(e), Pn(t), Le(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function et(e = !0) {
	Fn(null), Pn(null), Le(null), e && P?.deactivate();
}
function tt() {
	var e = V, t = e.b, n = P, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function nt(e) {
	var n = 2 | x;
	return V !== null && (V.f |= ne), {
		ctx: M,
		deps: null,
		effects: null,
		equals: we,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: V,
		ac: null
	};
}
var rt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function it(e, n, r) {
	let i = V;
	i === null && De();
	var a = void 0, o = Nt(t), s = !B, c = /* @__PURE__ */ new Set();
	return hn(() => {
		var t = V, n = y();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== E && n.reject(e);
			}).finally(et);
		} catch (e) {
			n.reject(e), et();
		}
		var r = P;
		if (s) {
			if (t.f & 32768) var l = tt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(rt);
			else for (let e of c.values()) e.reject(rt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== rt && (r.activate(), t ? (o.f |= se, zt(o, t)) : (o.f & 8388608 && (o.f ^= se), zt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), un(() => {
		for (let e of c) e.reject(rt);
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
function at(e) {
	let t = /* @__PURE__ */ nt(e);
	return Ln(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function ot(e) {
	let t = /* @__PURE__ */ nt(e);
	return t.equals = Ee, t;
}
function st(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) Sn(t[n]);
	}
}
function ct(e) {
	var n, r = V, i = e.parent;
	if (!jn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	Fn(i);
	try {
		st(e), n = Yn(e);
	} finally {
		Fn(r);
	}
	return n;
}
function lt(e) {
	var t = ct(e);
	!e.equals(t) && (e.wv = Kn(), (!P?.is_fork || e.deps === null) && (P === null ? e.v = t : (P.capture(e, t, !0), pt?.capture(e, t, !0)), e.deps === null)) ? N(e, b) : jn || (mt === null ? qe(e) : (ln() || P?.is_fork) && mt.set(e, t));
}
function ut(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Ze(() => {
		t.ac.abort(E), t.ac = null;
	}), t.fn !== null && (t.teardown = g), Qn(t, 0), bn(t));
}
function dt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && $n(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ft = null, P = null, pt = null, mt = null, ht = null, gt = !1, _t = !1, vt = null, yt = null, bt = 0, xt = 1, St = class e {
	id = xt++;
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
		ft === null ? ft = this : (ft.next = this, this.prev = ft), ft = this;
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
			for (var r of n.d) N(r, x), t(r);
			for (r of n.m) N(r, S), t(r);
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
		for (let [e, t] of this.#s) N(e, t), this.schedule(e);
		this.apply();
		for (var t = vt = [], n = [], r = yt = []; this.#a.length > 0;) {
			bt++ > 1e3 && (this.#b(), wt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw kt(e), this.#d() || this.discard(), t;
			}
		}
		if (P = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (vt = null, yt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Ot(e, t);
			r.length > 0 && P.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), pt = this, Et(n), Et(t), pt = null, this.#i?.resolve();
			var o = P;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (jt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= b;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= b : i & 4 ? t.push(r) : qn(r) && (i & 16 && this.add_dirty_reaction(r, S), $n(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), N(i, x), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), P = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Je(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), mt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		P = this;
	}
	deactivate() {
		P = null, mt = null;
	}
	flush() {
		try {
			_t = !0, P = this, this.#m();
		} finally {
			bt = 0, ht = null, vt = null, yt = null, _t = !1, P = null, mt = null, jt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(rt);
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
		this.#u || (this.#u = !0, We(() => {
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
		return (this.#i ??= y()).promise;
	}
	static ensure() {
		if (P === null) {
			let t = P = new e();
			!_t && !gt && We(() => {
				t.started || t.flush();
			});
		}
		return P;
	}
	apply() {
		mt = null;
	}
	schedule(e) {
		ht = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ft = e : t.prev = e, this.linked = !1;
		}
	}
};
function Ct(e) {
	var t = gt, n = pt;
	pt = null, gt = !0;
	try {
		var r;
		for (e && (Ct(), r = e());;) {
			if (Ge(), P === null) return r;
			P.flush();
		}
	} finally {
		gt = t, pt = n;
	}
}
function wt() {
	try {
		Me();
	} catch (e) {
		an(e, ht);
	}
}
var Tt = null;
function Et(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && qn(r) && (Tt = /* @__PURE__ */ new Set(), $n(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && wn(r), Tt?.size > 0)) {
				jt.clear();
				for (let e of Tt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Tt.has(n) && (Tt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || $n(n);
					}
				}
				Tt.clear();
			}
		}
		Tt = null;
	}
}
function Dt(e) {
	P.schedule(e);
}
function Ot(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), N(e, b);
		for (var n = e.first; n !== null;) Ot(n, t), n = n.next;
	}
}
function kt(e) {
	N(e, b);
	for (var t = e.first; t !== null;) kt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var At = /* @__PURE__ */ new Set(), jt = /* @__PURE__ */ new Map(), Mt = !1;
function Nt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: we,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Pt(e, t) {
	let n = Nt(e, t);
	return Ln(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Ft(e, t = !1, n = !0) {
	let r = Nt(e);
	return t || (r.equals = Ee), r;
}
function It(e, t, n = !1) {
	return B !== null && (!Nn || B.f & 131072) && Ve() && B.f & 4325394 && (In === null || !In.has(e)) && Fe(), zt(e, n ? F(t) : t, yt);
}
var Lt = null, Rt = 0;
function zt(e, t, n = null) {
	if (!e.equals(t)) {
		jn ? jt.set(e, t) : jt.has(e) || jt.set(e, e.v);
		var r = St.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ct(t), mt === null && qe(t);
		}
		e.wv = Kn(), Lt = null, Rt = 0, Ut(e, x, n), Lt = null, Ve() && V !== null && V.f & 1024 && !(V.f & 96) && (Bn === null ? Vn([e]) : Bn.push(e)), !r.is_fork && At.size > 0 && !Mt && Bt();
	}
	return t;
}
function Bt() {
	Mt = !1;
	for (let e of At) {
		e.f & 1024 && N(e, S);
		let t;
		try {
			t = qn(e);
		} catch {
			t = !0;
		}
		t && $n(e);
	}
	At.clear();
}
function Vt(e, t = 1) {
	var n = H(e), r = t === 1 ? n++ : n--;
	return It(e, n), r;
}
function Ht(e) {
	It(e, e.v + 1);
}
function Ut(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ve(), a = r.length;
		if (Rt += a, Rt > 1e5 && Lt === null && (Lt = /* @__PURE__ */ new Set()), Lt !== null) {
			if (Lt.has(e)) return;
			Lt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== V) {
				var l = (c & x) === 0;
				if (l && N(s, t), c & 131072) At.add(s);
				else if (c & 2) {
					var u = s;
					mt?.delete(u), Ut(u, S, n);
				} else if (l) {
					var d = s;
					c & 16 && Tt !== null && Tt.add(d), n === null ? Dt(d) : n.push(d);
				}
			}
		}
	}
}
function F(e) {
	if (typeof e != "object" || !e || ce in e || le in e) return e;
	let n = m(e);
	if (n !== f && n !== p) return e;
	var r = /* @__PURE__ */ new Map(), i = a(e), o = /* @__PURE__ */ Pt(0), s = null, c = Wn, l = (e) => {
		if (Wn === c) return e();
		var t = B, n = Wn;
		Pn(null), Gn(c);
		var r = e();
		return Pn(t), Gn(n), r;
	};
	return i && r.set("length", /* @__PURE__ */ Pt(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ne();
			var i = r.get(t);
			return i === void 0 ? l(() => {
				var e = /* @__PURE__ */ Pt(n.value, s);
				return r.set(t, e), e;
			}) : It(i, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var i = r.get(n);
			if (i === void 0) {
				if (n in e) {
					let e = l(() => /* @__PURE__ */ Pt(t, s));
					r.set(n, e), Ht(o);
				}
			} else It(i, t), Ht(o);
			return !0;
		},
		get(n, i, a) {
			if (i === ce) return e;
			var o = r.get(i), c = i in n;
			if (o === void 0 && (!c || u(n, i)?.writable) && (o = l(() => /* @__PURE__ */ Pt(F(c ? n[i] : t), s)), r.set(i, o)), o !== void 0) {
				var d = H(o);
				return d === t ? void 0 : d;
			}
			return Reflect.get(n, i, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var i = Reflect.getOwnPropertyDescriptor(e, n), a = r.get(n);
			if (a !== void 0) {
				var o = H(a);
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
			return (i !== void 0 || (V !== null || B !== null) && (!a || u(e, n)?.writable)) && (i === void 0 && (i = l(() => /* @__PURE__ */ Pt(a ? F(e[n]) : t, s)), r.set(n, i)), H(i) === t) ? !1 : a;
		},
		set(e, n, a, c) {
			var d = r.get(n), f = n in e;
			if (i && n === "length") for (var p = a; p < d.v; p += 1) {
				var m = r.get(p + "");
				m === void 0 ? p in e && (m = l(() => /* @__PURE__ */ Pt(t, s)), r.set(p + "", m)) : It(m, t);
			}
			if (d === void 0) (!f || u(e, n)?.writable) && (d = l(() => /* @__PURE__ */ Pt(void 0, s)), It(d, F(a)), r.set(n, d));
			else {
				f = d.v !== t;
				var h = l(() => F(a));
				It(d, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, a), !f) {
				if (i && typeof n == "string") {
					var _ = r.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && It(_, v + 1);
				}
				Ht(o);
			}
			return !0;
		},
		ownKeys(e) {
			H(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = r.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [i, a] of r) a.v !== t && !(i in e) && n.push(i);
			return n;
		},
		setPrototypeOf() {
			Pe();
		}
	});
}
var Wt, Gt, Kt, qt;
function Jt() {
	if (Wt === void 0) {
		Wt = window, Gt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Kt = u(t, "firstChild").get, qt = u(t, "nextSibling").get, h(e) && (e[fe] = void 0, e[de] = null, e[pe] = void 0, e.__e = void 0), h(n) && (n[me] = void 0);
	}
}
function Yt(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function I(e) {
	return Kt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Xt(e) {
	return qt.call(e);
}
function Zt(e, t) {
	if (!D) return /* @__PURE__ */ I(e);
	var n = /* @__PURE__ */ I(O);
	if (n === null) n = O.appendChild(Yt());
	else if (t && n.nodeType !== 3) {
		var r = Yt();
		return n?.before(r), k(r), r;
	}
	return t && nn(n), k(n), n;
}
function Qt(e, t = !1) {
	if (!D) {
		var n = /* @__PURE__ */ I(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Xt(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = Yt();
			return O?.before(r), k(r), r;
		}
		nn(O);
	}
	return O;
}
function L(e, t = !1) {
	if (!D) return /* @__PURE__ */ I(e);
	var n = Zt(e, t);
	return A(e), n;
}
function R(e, t = 1, n = !1) {
	let r = D ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Xt(r);
	if (!D) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = Yt();
			return r === null ? i?.after(a) : r.before(a), k(a), a;
		}
		nn(r);
	}
	return k(r), r;
}
function $t(e) {
	e.textContent = "";
}
function en() {
	return !1;
}
function tn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function nn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function rn(e) {
	var t = V;
	if (t === null) return B.f |= se, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	an(e, t);
}
function an(t, n) {
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
function on(e) {
	V === null && (B === null && je(e), Ae()), jn && ke(e);
}
function sn(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function cn(e, t) {
	var n = V;
	n !== null && n.f & 8192 && (e |= C);
	var r = {
		ctx: M,
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
	if (e & 4) vt === null ? St.ensure().schedule(r) : vt.push(r);
	else if (t !== null) {
		try {
			$n(r);
		} catch (e) {
			throw Sn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= T));
	}
	if (i !== null && (i.parent = n, n !== null && sn(i, n), B !== null && B.f & 2 && !(e & 64))) {
		var a = B;
		(a.effects ??= []).push(i);
	}
	return r;
}
function ln() {
	return B !== null && !Nn;
}
function un(e) {
	let t = cn(8, null);
	return N(t, b), t.teardown = e, t;
}
function dn(e) {
	on("$effect");
	var t = V.f;
	if (!B && t & 32 && M !== null && !M.i) {
		var n = M;
		(n.e ??= []).push(e);
	} else return fn(e);
}
function fn(e) {
	return cn(4 | re, e);
}
function pn(e) {
	St.ensure();
	let t = cn(64 | ne, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? Tn(t, () => {
			Sn(t), n(void 0);
		}) : (Sn(t), n(void 0));
	});
}
function mn(e) {
	return cn(4, e);
}
function hn(e) {
	return cn(oe | ne, e);
}
function gn(e, t = 0) {
	return cn(8 | t, e);
}
function z(e, t = [], n = [], r = []) {
	Qe(r, t, n, (t) => {
		cn(8, () => {
			e(...t.map(H));
		});
	});
}
function _n(e, t = 0) {
	return cn(16 | t, e);
}
function vn(e) {
	return cn(32 | ne, e);
}
function yn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = jn, r = B;
		Mn(!0), Pn(null);
		try {
			t.call(null);
		} catch (t) {
			an(t, e.parent);
		} finally {
			Mn(n), Pn(r);
		}
	}
}
function bn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Ze(() => {
			e.abort(E);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : Sn(n, t), n = r;
	}
}
function xn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || Sn(t), t = n;
	}
}
function Sn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (Cn(e.nodes.start, e.nodes.end), n = !0), e.f |= te, bn(e, t && !n), Qn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	yn(e), e.f ^= te, e.f |= w;
	var i = e.parent;
	i !== null && i.first !== null && wn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function Cn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Xt(e);
		e.remove(), e = n;
	}
}
function wn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function Tn(e, t, n = !0) {
	var r = [];
	e.f |= 256, En(e, r, !0);
	var i = () => {
		n && Sn(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function En(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= C;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				En(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Dn(e) {
	e.f &= -257, On(e, !0);
}
function On(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= C, e.f & 1024 || (N(e, x), St.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			On(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function kn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Xt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var An = null, jn = !1;
function Mn(e) {
	jn = e;
}
var B = null, Nn = !1;
function Pn(e) {
	B = e;
}
var V = null;
function Fn(e) {
	V = e;
}
var In = null;
function Ln(e) {
	B !== null && (B.f & 2097152 || B.f & 2) && (In ??= /* @__PURE__ */ new Set()).add(e);
}
var Rn = null, zn = 0, Bn = null;
function Vn(e) {
	Bn = e;
}
var Hn = 1, Un = 0, Wn = Un;
function Gn(e) {
	Wn = e;
}
function Kn() {
	return ++Hn;
}
function qn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (qn(a) && lt(a), a.wv > e.wv) return !0;
		}
		t & 512 && mt === null && N(e, b);
	}
	return !1;
}
function Jn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(In !== null && In.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Jn(a, t, !1) : t === a && (n ? N(a, x) : a.f & 1024 && N(a, S), Dt(a));
	}
}
function Yn(e) {
	var t = Rn, n = zn, r = Bn, i = B, a = In, o = M, s = Nn, c = Wn, l = e.f;
	Rn = null, zn = 0, Bn = null, B = l & 96 ? null : e, In = null, Le(e.ctx), Nn = !1, Wn = ++Un, e.ac !== null && (Ze(() => {
		e.ac.abort(E);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = Xn(e);
		if (Ve() && Bn !== null && !Nn && f !== null && !(e.f & 6146)) for (var p = 0; p < Bn.length; p++) Jn(Bn[p], e);
		if (i !== null && i !== e) {
			if (Un++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Un;
			if (t !== null) for (let e of t) e.rv = Un;
			Bn !== null && (r === null ? r = Bn : r.push(...Bn));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return Xn(e), rn(t);
	} finally {
		e.f ^= ae, Rn = t, zn = n, Bn = r, B = i, In = a, Le(o), Nn = s, Wn = c;
	}
}
function Xn(e) {
	var t = e.deps, n = P?.is_fork;
	if (Rn !== null) {
		var r;
		if (n || Qn(e, zn), t !== null && zn > 0) for (t.length = zn + Rn.length, r = 0; r < Rn.length; r++) t[zn + r] = Rn[r];
		else e.deps = t = Rn;
		if (ln() && e.f & 512) for (r = zn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && zn < t.length && (Qn(e, zn), t.length = zn);
	return t;
}
function Zn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var i = o.call(r, e);
		if (i !== -1) {
			var a = r.length - 1;
			a === 0 ? r = n.reactions = null : (r[i] = r[a], r.pop());
		}
	}
	if (r === null && n.f & 2 && (Rn === null || !s.call(Rn, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && qe(c), c.ac !== null && Ze(() => {
			c.ac.abort(E), c.ac = null, N(c, x);
		}), ut(c), Qn(c, 0);
	}
}
function Qn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Zn(e, n[r]);
}
function $n(e) {
	var t = e.f;
	if (!(t & 16384)) {
		N(e, b);
		var n = V;
		V = e;
		try {
			t & 16777232 ? xn(e) : bn(e), yn(e);
			var r = Yn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Hn;
		} finally {
			V = n;
		}
	}
}
function H(e) {
	var t = !!(e.f & 2);
	if (An?.add(e), B !== null && !Nn && !(V !== null && V.f & 16384) && (In === null || !In.has(e))) {
		var n = B.deps;
		if (B.f & 2097152) e.rv < Un && (e.rv = Un, Rn === null && n !== null && n[zn] === e ? zn++ : Rn === null ? Rn = [e] : Rn.push(e));
		else {
			B.deps ??= [], s.call(B.deps, e) || B.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [B] : s.call(r, B) || r.push(B);
		}
	}
	if (jn && jt.has(e)) return jt.get(e);
	if (t) {
		var i = e;
		if (jn) {
			var a = i.v;
			return (!(i.f & 1024) && i.reactions !== null || tr(i)) && (a = ct(i)), jt.set(i, a), a;
		}
		var o = !(i.f & 512) && !Nn && B !== null && !!(B.f & 512), c = (i.f & ee) === 0;
		qn(i) && (o && (i.f |= 512), lt(i)), o && !c && (dt(i), er(i));
	}
	if (mt?.has(e)) return mt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function er(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : s.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (dt(n), er(n));
	}
}
function tr(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (jt.has(t) || t.f & 2 && tr(t)) return !0;
	return !1;
}
function nr(e) {
	var t = Nn;
	try {
		return Nn = !0, e();
	} finally {
		Nn = t;
	}
}
function rr(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (ce in e) ir(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && ce in n && ir(n);
		}
	}
}
function ir(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			ir(e[n], t);
		} catch {}
		let n = m(e);
		if (n !== Object.prototype && n !== Array.prototype && n !== Map.prototype && n !== Set.prototype && n !== Date.prototype) {
			let t = d(n);
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
var ar = Symbol("events"), or = /* @__PURE__ */ new Set(), sr = /* @__PURE__ */ new Set();
function cr(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || mr.call(t, e), !e.cancelBubble) return Ze(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, We(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function lr(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = cr(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && un(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function ur(e, t, n) {
	(t[ar] ??= {})[e] = n;
}
function dr(e) {
	for (var t = 0; t < e.length; t++) or.add(e[t]);
	for (var n of sr) n(e);
}
var fr = null, pr = !1;
function mr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	fr = e, pr || (pr = !0, setTimeout(() => {
		pr = !1, fr = null;
	}));
	var o = 0, s = fr === e && e[ar];
	if (s) {
		var c = i.indexOf(s);
		if (c !== -1 && (t === document || t === window)) {
			e[ar] = t;
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
		var d = B, f = V;
		Pn(null), Fn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[ar]?.[r];
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
			e[ar] = t, delete e.currentTarget, Pn(d), Fn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var hr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function gr(e) {
	return hr?.createHTML(e) ?? e;
}
function _r(e) {
	var t = tn("template");
	return t.innerHTML = gr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function vr(e, t) {
	var n = V;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function U(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (D) return vr(O, null), O;
		i === void 0 && (i = _r(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ I(i)));
		var t = r || Gt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ I(t), s = t.lastChild;
			vr(o, s);
		} else vr(t, t);
		return t;
	};
}
function yr(e = "") {
	if (!D) {
		var t = Yt(e + "");
		return vr(t, t), t;
	}
	var n = O;
	return n.nodeType === 3 ? nn(n) : (n.before(n = Yt()), k(n)), vr(n, n), n;
}
function br() {
	if (D) return vr(O, null), O;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = Yt();
	return e.append(t, n), vr(t, n), e;
}
function W(e, t) {
	if (D) {
		var n = V;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), xe();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var xr = ["touchstart", "touchmove"];
function Sr(e) {
	return xr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Cr(e) {
	let t = 0, n = Nt(0), r;
	return () => {
		ln() && (H(n), gn(() => (t === 0 && (r = nr(() => e(() => Ht(n)))), t += 1, () => {
			We(() => {
				--t, t === 0 && (r?.(), r = void 0, Ht(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var wr = T | ne;
function Tr(e, t, n, r) {
	new Er(e, t, n, r);
}
var Er = class {
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
	#m = Cr(() => (this.#p = Nt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = V;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = V.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = _n(() => {
			if (D) {
				let e = Ce(this.#t);
				xe();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, wr), D && (this.#e = O);
	}
	#h() {
		try {
			this.#a = vn(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		We(r), t && (this.#s = vn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Ie(), this.#s !== null && Tn(this.#s, () => {
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
					an(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (w | te)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = vn(() => e(this.#e)), We(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = Yt(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return vn(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							an(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(P)) : this.#u === 0 && (this.#e.before(e), this.#c = null, Tn(this.#o, () => {
					this.#o = null;
				}), this.#x(P));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = vn(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				kn(this.#a, e);
				let t = this.#n.pending;
				this.#o = vn(() => t(this.#e));
			} else this.#x(P);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Je(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = V, n = B, r = M;
		Fn(this.#i), Pn(this.#i), Le(this.#i.ctx);
		try {
			return St.ensure(), e();
		} finally {
			Fn(t), Pn(n), Le(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && Tn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, We(() => {
			this.#d = !1, this.#p && zt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), H(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		P?.is_fork ? (this.#a && P.skip_effect(this.#a), this.#o && P.skip_effect(this.#o), this.#s && P.skip_effect(this.#s), P.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (Sn(this.#a), null), this.#o &&= (Sn(this.#o), null), this.#s &&= (Sn(this.#s), null), D && (k(this.#t), Se(), k(j()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return vn(() => {
						var r = V;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return an(e, this.#i.parent), null;
				}
			}));
		};
		We(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					an(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => an(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function G(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[me] ??= e.nodeValue) && (e[me] = n, e.nodeValue = `${n}`);
}
function Dr(e, t) {
	return kr(e, t);
}
var Or = /* @__PURE__ */ new Map();
function kr(t, { target: n, anchor: r, props: i = {}, events: a, context: o, intro: s = !0, transformError: l }) {
	Jt();
	var u = void 0, d = pn(() => {
		var s = r ?? n.appendChild(Yt());
		Tr(s, { pending: () => {} }, (n) => {
			Re({});
			var r = M;
			if (o && (r.c = o), a && (i.$$events = a), D && vr(n, null), u = t(n, i) || Be(), D && (V.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw ve(), e;
			ze();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Sr(r);
					for (let e of [n, document]) {
						var a = Or.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Or.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, mr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(c(or)), sr.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Or.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, mr), t.delete(e), t.size === 0 && Or.delete(r)) : t.set(e, i);
			}
			sr.delete(f), s !== r && s.parentNode?.removeChild(s);
		};
	});
	return Ar.set(u, d), u;
}
var Ar = /* @__PURE__ */ new WeakMap();
function jr(e, t) {
	let n = Ar.get(e);
	return n ? (Ar.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var Mr = class {
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
			if (n) Dn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Dn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (Sn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						kn(r, t), t.append(Yt()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else Sn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), Tn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (Sn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = P, r = en();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = Yt();
				i.append(a), this.#n.set(e, {
					effect: vn(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, vn(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else D && (this.anchor = O), this.#a(n);
	}
}, Nr = 0, Pr = 1, Fr = 2;
function Ir(e, n, r, i, a) {
	D && xe();
	var o = Ve(), s = t, c = o ? Nt(s) : /* @__PURE__ */ Ft(s, !1, !1), l = o ? Nt(s) : /* @__PURE__ */ Ft(s, !1, !1), u = new Mr(e);
	_n(() => {
		var t = P, o = n(), s = !1;
		let d = D && _(o) === (e.data === "[!");
		if (d && (k(j()), be(!1)), _(o)) {
			var f = $e(), p = !1;
			let e = (e) => {
				if (!s) {
					p = !0, f(!1), P === t && t.deactivate(), St.ensure();
					try {
						e();
					} finally {
						et(!1), gt || Ct();
					}
				}
			};
			o.then((t) => {
				e(() => {
					zt(c, t), u.ensure(Pr, i && ((e) => i(e, c)));
				});
			}, (t) => {
				e(() => {
					if (zt(l, t), u.ensure(Fr, a && ((e) => a(e, l))), !a) throw l.v;
				});
			}), D ? u.ensure(Nr, r) : We(() => {
				p || e(() => {
					u.ensure(Nr, r);
				});
			});
		} else zt(c, o), u.ensure(Pr, i && ((e) => i(e, c)));
		return d && be(!0), () => {
			s = !0;
		};
	});
}
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function Lr(e, t, n = !1) {
	var r;
	D && (r = O, xe());
	var i = new Mr(e), a = n ? T : 0;
	function o(e, t) {
		if (D) {
			var n = Ce(r);
			if (e !== parseInt(n.substring(1))) {
				var a = j();
				k(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	_n(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function Rr(e, t, n) {
	for (var r = [], i = t.length, a, o = t.length, s = 0; s < i; s++) {
		let n = t[s];
		Tn(n, () => {
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
			$t(d), d.append(u), e.items.clear();
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
		r?.has(a) ? (a.f |= ie, kn(a, document.createDocumentFragment())) : Sn(t[i], n);
	}
}
var Br;
function Vr(e, t, n, r, i, o = null) {
	var s = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		s = D ? k(/* @__PURE__ */ I(u)) : u.appendChild(Yt());
	}
	D && xe();
	var d = null, f = /* @__PURE__ */ ot(() => {
		var e = n();
		return a(e) ? e : e == null ? [] : c(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = H(f);
			_.fallback = d, Ur(_, n, s, t, r), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, Gr(d, null, s)) : Dn(d) : Tn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: _n(() => {
			var e = H(f), a = e.length;
			let c = !1;
			D && Ce(s) === "[!" != (a === 0) && (s = j(), k(s), be(!1), c = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = P, v = en(), y = 0; y < a; y += 1) {
				D && O.nodeType === 8 && O.data === "]" && (s = O, c = !0, be(!1));
				var b = e[y], x = r(b, y), S = m ? null : l.get(x);
				S ? (S.v && zt(S.v, b), S.i && zt(S.i, y), v && _.unskip_effect(S.e)) : (S = Wr(l, m ? s : Br ??= Yt(), b, x, y, i, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (a === 0 && o && !d && (m ? d = vn(() => o(s)) : (d = vn(() => o(Br ??= Yt())), d.f |= ie)), a > u.size && Oe("", "", ""), D && a > 0 && k(j()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			c && be(!0), H(f);
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
		if (_.f & 8192 && (Dn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
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
	a && We(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Wr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Nt(n) : /* @__PURE__ */ Ft(n, !1, !1) : null, l = o & 2 ? Nt(i) : null;
	return {
		v: c,
		i: l,
		e: vn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Gr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Xt(r);
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
		D && (l = k(/* @__PURE__ */ I(d)));
	}
	z(() => {
		var t = V;
		if (u === (u = n() ?? "")) D && xe();
		else if (a && !D) t.nodes = null, d.innerHTML = u, u !== "" && vr(/* @__PURE__ */ I(d), d.lastChild);
		else if (t.nodes !== null && (Cn(t.nodes.start, t.nodes.end), t.nodes = null), u !== "") {
			if (D) {
				for (var c = O.data, f = xe(), p = f; f !== null && (f.nodeType !== 8 || f.data !== "");) p = f, f = /* @__PURE__ */ Xt(f);
				if (f === null) throw ve(), e;
				vr(O, p), l = k(f);
			} else {
				var m = tn(o ? "svg" : s ? "math" : "template", o ? r : s ? i : void 0);
				m.innerHTML = u;
				var h = o || s ? m : m.content;
				if (vr(/* @__PURE__ */ I(h), h.lastChild), o || s) for (; /* @__PURE__ */ I(h);) l.before(/* @__PURE__ */ I(h));
				else l.before(h);
			}
		}
	});
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function Jr(e, t, n) {
	mn(() => {
		var r = nr(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			gn(() => {
				var e = n();
				rr(e), i && Te(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var Yr = [..." 	\n\r\f\xA0\v﻿"];
function Xr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Yr.includes(r[o - 1])) && (s === r.length || Yr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Zr(e, t, n, r, i, a) {
	var o = e[fe];
	if (D || o !== n || o === void 0) {
		var s = Xr(n, r, a);
		D && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[fe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Qr = Symbol("is custom element"), $r = Symbol("is html"), ei = ge ? "link" : "LINK", ti = ge ? "progress" : "PROGRESS";
function ni(e) {
	if (D) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					ai(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					ai(e, "checked", null), e.checked = r;
				}
			}
		};
		e[he] = n, We(n), Xe();
	}
}
function ri(e, t) {
	var n = oi(e);
	n.value !== (n.value = t ?? void 0) && (e.value !== t || t === 0 && e.nodeName === ti) && (e.value = t ?? "");
}
function ii(e, t) {
	var n = oi(e);
	n.checked !== (n.checked = t ?? void 0) && (e.checked = t);
}
function ai(e, t, n, r) {
	var i = oi(e);
	if (!(D && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === ei))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ue] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && ci(e).has(t) ? e[t] = n : (!D || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function oi(e) {
	return e[de] ??= {
		[Qr]: e.nodeName.includes("-"),
		[$r]: e.namespaceURI === n
	};
}
var si = /* @__PURE__ */ new Map();
function ci(e) {
	var t = e.getAttribute("is") || e.nodeName, n = si.get(t);
	if (n) return n;
	si.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = d(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = m(i);
	}
	return n;
}
// kit/plugin.svelte.js
function li(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ Pt(F({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return H(this.#e);
		}
		set props(e) {
			It(this.#e, e, !0);
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
			this.style.display = "contents", this.app ||= Dr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && jr(this.app), this.app = null;
		}
	});
}
function ui(e, t) {
	let n = F({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Vt(r);
		}
	}), r = /* @__PURE__ */ Pt(0);
	dn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Vt(r);
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
	return dn(() => {
		let n = e(), c = t();
		H(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function di(e) {
	let t = F({});
	return dn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
var fi = (e) => String(e?.message || e);
function pi(e, t) {
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
// node_modules/marked/lib/marked.esm.js
function mi() {
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
var hi = mi();
function gi(e) {
	hi = e;
}
var _i = { exec: () => null };
function vi(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function K(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(q.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var yi = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), q = {
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
	nextBulletRegex: vi((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: vi((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: vi((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: vi((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: vi((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: vi((e) => RegExp(`^ {0,${e}}(?:</?(?:${Pi})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: vi((e) => RegExp(`^ {0,${e}}>`))
}, bi = /^(?:[ \t]*(?:\n|$))+/, xi = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Si = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, Ci = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, wi = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, Ti = / {0,3}(?:[*+-]|\d{1,9}[.)])/, Ei = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, Di = K(Ei).replace(/bull/g, Ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Oi = K(Ei).replace(/bull/g, Ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), ki = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, Ai = /^[^\n]+/, ji = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Mi = K(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", ji).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), Ni = K(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, Ti).getRegex(), Pi = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Fi = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Ii = K("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Fi).replace("tag", Pi).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Li = (e) => K(ki).replace("hr", Ci).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Pi).getRegex(), Ri = Li(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), zi = Li(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Bi = {
	blockquote: K(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", zi).getRegex(),
	code: xi,
	def: Mi,
	fences: Si,
	heading: wi,
	hr: Ci,
	html: Ii,
	lheading: Di,
	list: Ni,
	newline: bi,
	paragraph: Ri,
	table: _i,
	text: Ai
}, Vi = K("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", Ci).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Pi).getRegex(), Hi = {
	...Bi,
	lheading: Oi,
	table: Vi,
	paragraph: K(ki).replace("hr", Ci).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Vi).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Pi).getRegex()
}, Ui = {
	...Bi,
	html: K("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Fi).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: _i,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: K(ki).replace("hr", Ci).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Di).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Wi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Gi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Ki = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, qi = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, Ji = /[\p{P}\p{S}]/u, Yi = /[\s\p{P}\p{S}]/u, Xi = /[^\s\p{P}\p{S}]/u, Zi = K(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Yi).getRegex(), Qi = /[\p{Pi}\p{Ps}"']/u, $i = /(?!~)[\p{P}\p{S}]/u, ea = /(?!~)[\s\p{P}\p{S}]/u, ta = /(?:[^\s\p{P}\p{S}]|~)/u, na = K(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", yi ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), ra = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, ia = K(ra, "u").replace(/punct/g, Ji).getRegex(), aa = K(ra, "u").replace(/punct/g, $i).getRegex(), oa = K(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, Qi).replace(/punct/g, Ji).getRegex(), sa = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", ca = K(sa, "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), la = K(sa, "gu").replace(/notPunctSpace/g, ta).replace(/punctSpace/g, ea).replace(/punct/g, $i).getRegex(), ua = K("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), da = K("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), fa = K("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), pa = K(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, Ji).getRegex(), ma = K("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Xi).replace(/punctSpace/g, Yi).replace(/punct/g, Ji).getRegex(), ha = K(/\\(punct)/, "gu").replace(/punct/g, Ji).getRegex(), ga = K(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, Ji).getRegex(), _a = K(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), va = K(Fi).replace("(?:-->|$)", "-->").getRegex(), ya = K("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", va).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), ba = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, xa = K(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", ba).getRegex(), Sa = K(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", xa).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), Ca = K(/^!?\[(label)\]\[(ref)\]/).replace("label", xa).replace("ref", ji).getRegex(), wa = K(/^!?\[(ref)\](?:\[\])?/).replace("ref", ji).getRegex(), Ta = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, Ea = K(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", ba).getRegex(), Da = K("reflink|nolink(?!\\()", "g").replace("reflink", K(/^!?\[(label)\]\[(ref)\]/).replace("label", Ea).replace("ref", Ta).getRegex()).replace("nolink", K(/^!?\[(ref)\](?:\[\])?/).replace("ref", Ta).getRegex()).getRegex(), Oa = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, ka = K(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), Aa = {
	_backpedal: _i,
	anyPunctuation: ha,
	anyPunctuationOutsideCode: ga,
	autolink: _a,
	blockSkip: na,
	br: Ki,
	code: Gi,
	del: _i,
	delLDelim: _i,
	delRDelim: _i,
	emStrongLDelim: ia,
	emStrongRDelimAst: ca,
	emStrongRDelimUnd: da,
	escape: Wi,
	link: Sa,
	nolink: wa,
	punctuation: Zi,
	reflink: Ca,
	reflinkSearch: Da,
	tag: ya,
	text: qi,
	url: _i
}, ja = {
	...Aa,
	emStrongLDelim: oa,
	emStrongRDelimAst: ua,
	emStrongRDelimUnd: fa,
	link: K(/^!?\[(label)\]\((.*?)\)/).replace("label", xa).getRegex(),
	reflink: K(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", xa).getRegex()
}, Ma = {
	...Aa,
	emStrongRDelimAst: la,
	emStrongLDelim: aa,
	delLDelim: pa,
	delRDelim: ma,
	url: K(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", ka).replace("protocol", Oa).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: K(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", Oa).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", yi ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, Na = {
	...Ma,
	br: K(Ki).replace("{2,}", "*").getRegex(),
	text: K(Ma.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, Pa = {
	normal: Bi,
	gfm: Hi,
	pedantic: Ui
}, Fa = {
	normal: Aa,
	gfm: Ma,
	breaks: Na,
	pedantic: ja
}, Ia = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, La = (e) => Ia[e];
function Ra(e, t) {
	if (t) {
		if (q.escapeTest.test(e)) return e.replace(q.escapeReplace, La);
	} else if (q.escapeTestNoEncode.test(e)) return e.replace(q.escapeReplaceNoEncode, La);
	return e;
}
function za(e) {
	return q.numericCharacterReference.test(e) ? e.replace(q.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Ba(e) {
	try {
		e = encodeURI(e).replace(q.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Va(e, t) {
	let n = e.replace(q.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(q.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(q.slashPipe, "|");
	return n;
}
function Ha(e, t, n) {
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
function Ua(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && q.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Wa(e) {
	return e.trim().replace(q.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Ga(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Ka = /* @__PURE__ */ new WeakMap();
function qa(e) {
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
function Ja(e, t) {
	let n = qa(t), r = Ka.get(e);
	r ? r.push(n) : Ka.set(e, [n]);
}
function Ya(e) {
	let t = Ka.get(e);
	t && (t.pop(), t.length === 0 && Ka.delete(e));
}
function Xa(e, t, n) {
	let r = Ka.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : qa(t).lastPossibleEmStrongCloserIndex[n];
}
function Za(e, t) {
	let n = Ka.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : qa(t).lastPossibleDelCloserIndex;
}
function Qa(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function $a(e, t, n, r, i) {
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
function eo(e, t, n) {
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
function to(e, t, n, r) {
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
var no = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || hi;
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
			let e = this.options.pedantic ? t[0] : Ua(t[0]);
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
			let e = t[0], n = eo(e, t[3] || "", this.rules);
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
				let t = Ha(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: Ha(t[0], "\n"),
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
			raw: Ha(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = Ha(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + Qa(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? Qa(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => Qa(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => Qa(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = Ua(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && Ga(t[2], "()") !== -1) return;
			let e = Wa(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: Ha(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Va(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: Ha(t[0], "\n"),
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
			for (let e of i) a.rows.push(Va(e, a.header.length).map((e, t) => ({
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
				raw: Ha(t[0], "\n"),
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
			if (!this.options.pedantic && to(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = Ha(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Ga(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = Ha(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), $a(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && to(e, n[1], r, this.rules)) return;
			let i = t[Wa(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return $a(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Xa(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Za(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : za(t[0]),
				escaped: e
			};
		}
	}
}, ro = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || hi, this.options.tokenizer = this.options.tokenizer || new no(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: q,
			block: Pa.normal,
			inline: Fa.normal
		};
		this.options.pedantic ? (t.block = Pa.pedantic, t.inline = Fa.pedantic) : this.options.gfm && (t.block = Pa.gfm, t.inline = this.options.breaks ? Fa.breaks : Fa.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: Pa,
			inline: Fa
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(q.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(q.tabCharGlobal, "    ").replace(q.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Wa(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== no.prototype.emStrong || this.tokenizer.del !== no.prototype.del ? this.#n(e) : e;
		Ja(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Ya(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, Wa(t.slice(r + 1, -1)))) return t;
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
}, io = class {
	options;
	parser;
	constructor(e) {
		this.options = e || hi;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(q.notSpaceStart)?.[0], i = e ? e.replace(q.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Ra(r) + "\">" + (n ? i : Ra(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Ra(i, !0)) + "</code></pre>\n";
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
		return `<code>${Ra(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Ra(n, !0) : this.parser.parseInline(r), o = Ba(e);
		if (o === null) return a;
		e = Ra(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Ra(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Ba(e);
		if (i === null) return Ra(n);
		e = i;
		let a = `<img src="${Ra(e)}" alt="${Ra(n)}"`;
		return t && (a += ` title="${Ra(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Ra(e.text);
	}
}, ao = class {
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
}, oo = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || hi, this.options.renderer = this.options.renderer || new io(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new ao(), this.textRenderer.parser = this;
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
}, so = class {
	options;
	block;
	constructor(e) {
		this.options = e || hi;
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
		return e ? ro.lex : ro.lexInline;
	}
	provideParser(e = this.block) {
		return e ? oo.parse : oo.parseInline;
	}
};
function co(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var lo = new class {
	defaults = mi();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = oo;
	Renderer = io;
	TextRenderer = ao;
	Lexer = ro;
	Tokenizer = no;
	Hooks = so;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (co(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) co(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) co(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				co(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					co(n, this.walkTokens(i, t));
				}) : e.tokens && co(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new io(this.defaults);
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
				let t = this.defaults.tokenizer || new no(this.defaults);
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
				let t = this.defaults.hooks || new so();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = so.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && so.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return co(n, r.call(this, e)), t && co(n, t.call(this, e)), n;
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
		return ro.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return oo.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? ro.lex : ro.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? oo.parse : oo.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? ro.lex : ro.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? oo.parse : oo.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Ra(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function J(e, t) {
	return lo.parse(e, t);
}
J.options = J.setOptions = function(e) {
	return lo.setOptions(e), J.defaults = lo.defaults, gi(J.defaults), J;
}, J.getDefaults = mi, J.defaults = hi;
function uo(...e) {
	return lo.use(...e), J.defaults = lo.defaults, gi(J.defaults), J;
}
J.use = uo, J.walkTokens = function(e, t) {
	return lo.walkTokens(e, t);
}, J.parseInline = lo.parseInline, J.Parser = oo, J.parser = oo.parse, J.Renderer = io, J.TextRenderer = ao, J.Lexer = ro, J.lexer = ro.lex, J.Tokenizer = no, J.Hooks = so, J.parse = J, J.options, J.setOptions, J.walkTokens, J.parseInline, oo.parse, ro.lex;
// node_modules/dompurify/dist/purify.es.mjs
function fo(e, t) {
	this.v = e, this.k = t;
}
function po(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function mo(e) {
	if (Array.isArray(e)) return e;
}
function ho(e, t) {
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
function go() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function _o(e, t) {
	return mo(e) || ho(e, t) || vo(e, t) || go();
}
function vo(e, t) {
	if (e) {
		if (typeof e == "string") return po(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? po(e, t) : void 0;
	}
}
function yo(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof fo;
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
yo.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, yo.prototype.next = function(e) {
	return this._invoke("next", e);
}, yo.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, yo.prototype.return = function(e) {
	return this._invoke("return", e);
};
var bo = Object.entries, xo = Object.setPrototypeOf, So = Object.isFrozen, Co = Object.getPrototypeOf, wo = Object.getOwnPropertyDescriptor, Y = Object.freeze, To = Object.seal, Eo = Object.create, Do = typeof Reflect < "u" && Reflect, Oo = Do.apply, ko = Do.construct;
Y ||= function(e) {
	return e;
}, To ||= function(e) {
	return e;
}, Oo ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, ko ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Ao = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var jo = Z(Array.prototype.lastIndexOf), Mo = Z(Array.prototype.pop), No = Z(Array.prototype.push), Po = Z(Array.prototype.slice), Fo = Z(Array.prototype.splice), Io = Array.isArray, Lo = $o(Set), Ro = Z(Set.prototype.add), zo = Z(Set.prototype.has), Bo = Z(String.prototype.toLowerCase), Vo = Z(String.prototype.toString), Ho = Z(String.prototype.match), Uo = Z(String.prototype.replace), Wo = Z(String.prototype.indexOf), Go = Z(String.prototype.trim), Ko = Z(Number.prototype.toString), qo = Z(Boolean.prototype.toString), Jo = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Yo = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Xo = Z(Object.prototype.hasOwnProperty), Zo = Z(Object.prototype.toString), X = Z(RegExp.prototype.test), Qo = $o(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Oo(e, t, n);
	};
}
function $o(e) {
	return function() {
		return ko(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : Bo;
	if (xo && xo(e, null), !Io(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (So(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function es(e) {
	for (let t = 0; t < e.length; t++) Xo(e, t) || (e[t] = null);
	return e;
}
function ts(e) {
	let t = Eo(null);
	for (let r of bo(e)) {
		var n = _o(r, 2);
		let i = n[0], a = n[1];
		Xo(e, i) && (t[i] = Io(a) ? es(a) : a && typeof a == "object" && a.constructor === Object ? ts(a) : a);
	}
	return t;
}
function ns(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return Ko(e);
		case "boolean": return qo(e);
		case "bigint": return Jo ? Jo(e) : "0";
		case "symbol": return Yo ? Yo(e) : "Symbol()";
		case "undefined": return Zo(e);
		case "function":
		case "object": {
			if (e === null) return Zo(e);
			let t = e, n = rs(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : Zo(e);
			}
			return Zo(e);
		}
		default: return Zo(e);
	}
}
function rs(e, t) {
	for (; e !== null;) {
		let n = wo(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Co(e);
	}
	function n() {
		return null;
	}
	return n;
}
function is(e, t) {
	for (; e !== null;) {
		let n = wo(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = Co(e);
	}
	return null;
}
function as(e) {
	try {
		return X(e, ""), !0;
	} catch {
		return !1;
	}
}
var os = Y(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), ss = Y(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), cs = Y([
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
]), ls = Y([
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
]), us = Y(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), ds = Y([
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
]), fs = Y(["#text"]), ps = Y(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), ms = Y(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), hs = Y(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), gs = Y([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), _s = To(/{{[\w\W]*|^[\w\W]*}}/g), vs = To(/<%[\w\W]*|^[\w\W]*%>/g), ys = To(/\${[\w\W]*/g), bs = To(/^data-[\-\w.\u00B7-\uFFFF]+$/), xs = To(/^aria-[\-\w]+$/), Ss = To(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), Cs = To(/^(?:\w+script|data):/i), ws = To(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Ts = To(/^html$/i), Es = To(/^[a-z][.\w]*(-[.\w]+)+$/i), Ds = To(/<[/\w!]/g), Os = To(/<[/\w]/g), ks = To(/<\/no(script|embed|frames)/i), As = To(/\/>/i), $ = {
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
}, js = Y(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), Ms = function() {
	return typeof window > "u" ? null : window;
}, Ns = function(e, t) {
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
}, Ps = function() {
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
}, Fs = function(e, t, n, r) {
	return Xo(e, t) && Io(e[t]) ? Q(r.base ? ts(r.base) : {}, e[t], r.transform) : n;
}, Is = function(e, t, n) {
	let r = Xo(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? ts(r) : n();
};
function Ls() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : Ms(), t = (e) => Ls(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = rs(d, "cloneNode"), p = rs(d, "remove"), m = rs(d, "removeAttributeNode"), h = rs(d, "nextSibling"), g = rs(d, "childNodes"), _ = rs(d, "parentNode"), v = rs(d, "shadowRoot"), y = rs(d, "attributes"), b = o && o.prototype ? rs(o.prototype, "nodeType") : null, x = o && o.prototype ? rs(o.prototype, "nodeName") : null, S = o && o.prototype ? rs(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = is(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, ne = "", re, ie = !1, ae = 0, oe = function() {
		if (ae > 0) throw Qo("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
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
		return ie ||= (re = Ns(u, i), !0), re;
	}, ue = n, de = ue.implementation, fe = ue.createNodeIterator, pe = ue.createDocumentFragment, me = ue.getElementsByTagName, he = r.importNode, E = Ps();
	t.isSupported = typeof bo == "function" && typeof _ == "function" && de && de.createHTMLDocument !== void 0;
	let ge = _s, _e = vs, ve = ys, ye = bs, D = xs, be = Cs, O = ws, k = Es, xe = Ss, A = null, Se = Q({}, [
		...os,
		...ss,
		...cs,
		...us,
		...fs
	]), j = null, Ce = Q({}, [
		...ps,
		...ms,
		...hs,
		...gs
	]), we = Object.seal(Eo(null, {
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
	})), Te = null, Ee = null, De = Object.seal(Eo(null, {
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
	})), Oe = !0, ke = !0, Ae = !1, je = !0, Me = !1, Ne = !0, Pe = !1, Fe = !1, Ie = null, M = null, Le = !1, Re = !1, ze = !1, Be = !1, Ve = !0, He = !1, Ue = "user-content-", We = !0, Ge = !1, Ke = null, N = {}, qe = null, Je = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ye = null, Xe = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), Ze = null, Qe = Q({}, [
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
	]), $e = "http://www.w3.org/1998/Math/MathML", et = "http://www.w3.org/2000/svg", tt = "http://www.w3.org/1999/xhtml", nt = tt, rt = !1, it = null, at = Q({}, [
		$e,
		et,
		tt
	], Vo), ot = Y([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), st = Q({}, ot), ct = Y(["annotation-xml"]), lt = Q({}, ct), ut = Q({}, [
		"title",
		"style",
		"font",
		"a",
		"script"
	]), dt = null, ft = ["application/xhtml+xml", "text/html"], P = null, pt = null, mt = n.createElement("form"), ht = function(e) {
		return e instanceof RegExp || e instanceof Function;
	}, gt = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		if (pt && pt === e) return;
		e && typeof e == "object" || (e = {}), e = ts(e), dt = ft.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, P = dt === "application/xhtml+xml" ? Vo : Bo, A = Fs(e, "ALLOWED_TAGS", Se, { transform: P }), j = Fs(e, "ALLOWED_ATTR", Ce, { transform: P }), it = Fs(e, "ALLOWED_NAMESPACES", at, { transform: Vo }), Ze = Fs(e, "ADD_URI_SAFE_ATTR", Qe, {
			transform: P,
			base: Qe
		}), Ye = Fs(e, "ADD_DATA_URI_TAGS", Xe, {
			transform: P,
			base: Xe
		}), qe = Fs(e, "FORBID_CONTENTS", Je, { transform: P }), Te = Fs(e, "FORBID_TAGS", ts({}), { transform: P }), Ee = Fs(e, "FORBID_ATTR", ts({}), { transform: P }), N = Xo(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? ts(e.USE_PROFILES) : e.USE_PROFILES : !1, Oe = e.ALLOW_ARIA_ATTR !== !1, ke = e.ALLOW_DATA_ATTR !== !1, Ae = e.ALLOW_UNKNOWN_PROTOCOLS || !1, je = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Me = e.SAFE_FOR_TEMPLATES || !1, Ne = e.SAFE_FOR_XML !== !1, Pe = e.WHOLE_DOCUMENT || !1, Re = e.RETURN_DOM || !1, ze = e.RETURN_DOM_FRAGMENT || !1, Be = e.RETURN_TRUSTED_TYPE || !1, Le = e.FORCE_BODY || !1, Ve = e.SANITIZE_DOM !== !1, He = e.SANITIZE_NAMED_PROPS || !1, We = e.KEEP_CONTENT !== !1, Ge = e.IN_PLACE || !1, xe = as(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : Ss, nt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : tt, st = Is(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, ot)), lt = Is(e, "HTML_INTEGRATION_POINTS", () => Q({}, ct));
		let t = Is(e, "CUSTOM_ELEMENT_HANDLING", () => Eo(null));
		if (we = Eo(null), Xo(t, "tagNameCheck") && ht(t.tagNameCheck) && (we.tagNameCheck = t.tagNameCheck), Xo(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (we.attributeNameCheck = t.attributeNameCheck), Xo(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (we.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), To(we), Me && (ke = !1), ze && (Re = !0), N && (A = Q({}, fs), j = Eo(null), N.html === !0 && (Q(A, os), Q(j, ps)), N.svg === !0 && (Q(A, ss), Q(j, ms), Q(j, gs)), N.svgFilters === !0 && (Q(A, cs), Q(j, ms), Q(j, gs)), N.mathMl === !0 && (Q(A, us), Q(j, hs), Q(j, gs))), De.tagCheck = null, De.attributeCheck = null, Xo(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? De.tagCheck = e.ADD_TAGS : Io(e.ADD_TAGS) && (A === Se && (A = ts(A)), Q(A, e.ADD_TAGS, P))), Xo(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? De.attributeCheck = e.ADD_ATTR : Io(e.ADD_ATTR) && (j === Ce && (j = ts(j)), Q(j, e.ADD_ATTR, P))), Xo(e, "ADD_FORBID_CONTENTS") && Io(e.ADD_FORBID_CONTENTS) && (qe === Je && (qe = ts(qe)), Q(qe, e.ADD_FORBID_CONTENTS, P)), We && (A["#text"] = !0), Pe && Q(A, [
			"html",
			"head",
			"body"
		]), A.table && (Q(A, ["tbody"]), delete Te.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw Qo("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw Qo("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = T;
			T = e.TRUSTED_TYPES_POLICY;
			try {
				ne = se("");
			} catch (e) {
				throw T = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (T = void 0, ne = "") : (T === void 0 && (T = le()), T && typeof ne == "string" && (ne = se("")));
		Y && Y(e), pt = e;
	}, _t = Q({}, [
		...ss,
		...cs,
		...ls
	]), vt = Q({}, [...us, ...ds]), yt = function(e, t, n) {
		return t.namespaceURI === tt ? e === "svg" : t.namespaceURI === $e ? e === "svg" && (n === "annotation-xml" || st[n]) : !!_t[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === tt ? e === "math" : t.namespaceURI === et ? e === "math" && lt[n] : !!vt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === et && !lt[n] || t.namespaceURI === $e && !st[n] ? !1 : !vt[e] && (ut[e] || !_t[e]);
	}, St = function(e) {
		let t = te(e), n = Bo(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = Bo(w(r))) : (i = { namespaceURI: nt }, a = "template"), it[t] ? t === et ? yt(n, i, a) : t === $e ? bt(n, i, a) : t === tt ? xt(n, i, a) : !!(dt === "application/xhtml+xml" && it[t]) : !1;
	}, Ct = function(e) {
		No(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw Qo("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
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
		t && Ao(Po(t), (e) => {
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
		No(t.removed, {
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
			if (Re || ze) try {
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
			typeof i != "string" || j[P(i)] || wt(e, r, i);
		}
	}, Ot = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				F(t) && r.push(t);
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
		let n = Lo();
		Ao(At(e), (e) => {
			C(e) === $.element && Ro(n, e);
		}), Ao(t, (e) => {
			C(e) === $.element && !zo(n, e) && Dt(e);
		});
	}, Mt = function(e, n) {
		let r = !1;
		return Ao(t.removed, (t) => {
			t.element && (r ||= t.element === e, kt(t.element));
		}), jt(e, n), r;
	}, Nt = function(e, t) {
		let n = Ke, r = null, i = () => (r || (r = Lo(), Ao(t, (e) => {
			Ro(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = At(e), a = [e], o = !1;
			if (Ao(r, (e) => {
				C(e) === $.documentFragment ? No(a, e) : !o && !zo(n, e) && (o = zo(i(), e));
			}), !o) return;
			if (t >= 8) throw Qo("the tree kept changing during in-place sanitization; refusing to return it");
			Ao(a, (t, n) => {
				Pt(t, n > 0, e, i());
			});
		}
	}, Pt = function(e, t, n, r) {
		let i = zt(e), a;
		for (; a = i.nextNode();) zo(r, a) && !zo(Ke, a) && (t && Gt(E.uponSanitizeShadowNode, a, null), Xt(a, n), tn(a, n), F(a.content) && nn(a.content));
	}, Ft = function(e) {
		Ke && Ro(Ke, e);
	}, It = function(e, t) {
		return Ne ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Lt = function(e) {
		Ne && Ot(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && X(Os, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = P(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && It("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, Rt = function(e) {
		let t = null, r = null;
		if (Le) e = "<remove></remove>" + e;
		else {
			let t = Ho(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		dt === "application/xhtml+xml" && nt === tt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = T ? se(e) : e;
		if (nt === tt) try {
			t = new l().parseFromString(i, dt);
		} catch {}
		if (!t || !t.documentElement) {
			t = de.createDocument(nt, "template", null);
			try {
				t.documentElement.innerHTML = rt ? ne : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), nt === tt ? me.call(t, Pe ? "html" : "body")[0] : Pe ? t.documentElement : a;
	}, zt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return fe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Bt = function(e) {
		return e = Uo(e, ge, " "), e = Uo(e, _e, " "), e = Uo(e, ve, " "), e;
	}, Vt = function(e) {
		Ht(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && Ao(t, (e) => {
			F(e.content) && Vt(e.content);
		});
	}, Ht = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = fe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Bt(r.data), r = n.nextNode();
	}, Ut = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || P(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, F = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, Wt = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function Gt(e, n, r) {
		e.length !== 0 && Ao(e, (e) => {
			e.call(t, n, r, pt);
		});
	}
	let Kt = function(e, t) {
		return !!(Ne && e.hasChildNodes() && !Wt(e.firstElementChild) && X(Ds, e.textContent) && X(Ds, e.innerHTML) || Ne && js[t] && C(e) === $.element && te(e) === tt && (Wt(e.firstElementChild) || typeof e.textContent == "string" && X(Ds, e.textContent)) || e.nodeType === $.processingInstruction || Ne && e.nodeType === $.comment && X(Os, e.data));
	}, qt = function(e, t) {
		return e instanceof RegExp ? X(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Jt = function(e, t, n) {
		if (!Te[t] && L(t) && qt(we.tagNameCheck, t)) return !1;
		if (We && !qe[t] && !qe[Bo(t)]) {
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
	}, Yt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? ts(t) : t;
	}, I = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ge && kt(e), !0);
	}, Xt = function(e, n) {
		if (Ft(e), Gt(E.beforeSanitizeElements, e, null), I(e, n)) return !0;
		if (Ut(e)) return Ct(e), !0;
		let r = P(w(e));
		if (A = Yt(E.uponSanitizeElement, A, Se, Ie), Gt(E.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: A
		}), I(e, n)) return !0;
		if (Kt(e, r)) return Ct(e), !0;
		if (Te[r] || !(De.tagCheck instanceof Function && De.tagCheck(r)) && !A[r]) {
			let t = Jt(e, r, n);
			return t === !1 && (Gt(E.afterSanitizeElements, e, null), I(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !St(e) || (r === "noscript" || r === "noembed" || r === "noframes") && X(ks, e.innerHTML)) return Ct(e), !0;
		if (Me && e.nodeType === $.text) {
			let n = Bt(e.textContent);
			e.textContent !== n && (No(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Gt(E.afterSanitizeElements, e, null), I(e, n);
	}, Zt = function(e, t, r) {
		if (Ee[t] || It(t, e) || Ve && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = j[t] || De.attributeCheck instanceof Function && De.attributeCheck(t, e);
		return ke && X(ye, t) || Oe && X(D, t) ? !0 : i ? Ze[t] || X(xe, Uo(r, O, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && Wo(r, "data:") === 0 && Ye[e] || Ae && !X(be, Uo(r, O, "")) ? !0 : !r : L(e) && qt(we.tagNameCheck, e) && qt(we.attributeNameCheck, t, e) || t === "is" && we.allowCustomizedBuiltInElements && qt(we.tagNameCheck, r);
	}, Qt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), L = function(e) {
		return !Qt[Bo(e)] && X(k, e);
	}, R = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return se(r);
			case "TrustedScriptURL": return ce(r);
		}
		return r;
	}, $t = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Ut(e) || (Ct(e), !1);
		} catch {
			return Et(t, e), !1;
		}
	}, en = function(e, t) {
		return Ne && X(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : Bo(e) === "attributename" && Ho(t, "href") !== null;
	}, tn = function(e, n) {
		if (Gt(E.beforeSanitizeAttributes, e, null), I(e, n)) return;
		let r = e.attributes;
		if (!r || Ut(e)) return;
		j = Yt(E.uponSanitizeAttribute, j, Ce, M);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: j,
			forceKeepAttr: void 0
		}, a = r.length, o = P(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = P(s), d = l, f = s === "value" ? d : Go(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Gt(E.uponSanitizeAttribute, e, i), f = i.attrValue, He && (u === "id" || u === "name") && Wo(f, Ue) !== 0 && (Et(s, e, n), f = Ue + f, p = !0), en(s, f) ? Et(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !je && X(As, f) ? Et(s, e, n) : (Me && (f = Bt(f)), Zt(o, u, f) ? (f = R(o, u, c, f), f !== d && $t(e, s, c, f) && p && Mo(t.removed)) : Et(s, e, n)) : Et(s, e, n));
		}
		Gt(E.afterSanitizeAttributes, e, null), I(e, n);
	}, nn = function(e) {
		let t = null, n = zt(e);
		for (Gt(E.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Gt(E.uponSanitizeShadowNode, t, null), Xt(t, e), tn(t, e), F(t.content) && nn(t.content), C(t) === $.element) {
			let e = v(t);
			F(e) && (rn(e), nn(e));
		}
		Me && Ht(e), Gt(E.afterSanitizeShadowDOM, e, null);
	}, rn = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				nn(e.shadow);
				continue;
			}
			let n = e.node, r = C(n) === $.element, i = g(n);
			if (i) for (let e = i.length - 1; e >= 0; --e) t.push({
				node: i[e],
				shadow: null
			});
			if (r) {
				let e = x ? x(n) : null;
				if (typeof e == "string" && P(e) === "template") {
					let e = n.content;
					F(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				F(e) && t.push({
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
		if (rt = !e, rt && (e = "<!-->"), typeof e != "string" && !Wt(e) && (e = ns(e), typeof e != "string")) throw Qo("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Fe ? (A = Ie, j = M) : gt(n), (E.uponSanitizeElement.length > 0 || E.uponSanitizeAttribute.length > 0) && (A = ts(A)), E.uponSanitizeAttribute.length > 0 && (j = ts(j)), t.removed = [];
		let c = Ge && typeof e != "string" && Wt(e), l = [], u = Ke;
		if (c) {
			Lt(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = P(t);
				if (!A[n] || Te[n]) throw Tt(e), Qo("root node is forbidden and cannot be sanitized in-place");
			}
			if (Ut(e)) throw Tt(e), Qo("root node is clobbered and cannot be sanitized in-place");
			l = At(e), Ke = Lo();
		} else if (Wt(e)) i = Rt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), rn(i);
		else {
			if (!Re && !Me && !Pe && e.indexOf("<") === -1) return T && Be ? se(e) : e;
			if (i = Rt(e), !i) return Re ? null : Be ? ne : "";
		}
		i && Le && Ct(i.firstChild);
		let d = c ? e : i;
		try {
			c && rn(e);
			let t = zt(d);
			for (; o = t.nextNode();) Xt(o, d), tn(o, d), F(o.content) && nn(o.content);
			c && Nt(e, l);
		} catch (t) {
			throw c && (Tt(e), Mt(e, l)), t;
		} finally {
			Ke = u;
		}
		if (c) {
			if (Mt(e, l)) throw Qo("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Me && Vt(e), e;
		}
		if (Re) {
			if (Me && Vt(i), ze) for (s = pe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (j.shadowroot || j.shadowrootmode) && (s = he.call(r, s, !0)), s;
		}
		let f = Pe ? i.outerHTML : i.innerHTML;
		return Pe && A["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && X(Ts, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Me && (f = Bt(f)), T && Be ? se(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Fe = !0, Ie = A, M = j;
	}, t.clearConfig = function() {
		pt = null, Fe = !1, Ie = null, M = null, T = re, ne = "";
	}, t.isValidAttribute = function(e, t, n) {
		pt || gt({});
		let r = P(e), i = P(t);
		return typeof n == "string" && (en(t, n) || !je && X(As, n)) ? !1 : Zt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Xo(E, e) && No(E[e], t);
	}, t.removeHook = function(e, t) {
		if (Xo(E, e)) {
			if (t !== void 0) {
				let n = jo(E[e], t);
				return n === -1 ? void 0 : Fo(E[e], n, 1)[0];
			}
			return Mo(E[e]);
		}
	}, t.removeHooks = function(e) {
		Xo(E, e) && (E[e] = []);
	}, t.removeAllHooks = function() {
		E = Ps();
	}, t;
}
var Rs = Ls();
// kit/format.js
function zs(e) {
	let t = J.parse(e || "", {
		async: !1,
		gfm: !0
	}), n = Rs.sanitize(t, {
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
			"math"
		],
		FORBID_ATTR: ["style"]
	}), r = new DOMParser().parseFromString(n, "text/html");
	for (let e of r.querySelectorAll("a")) Hs(e.href) || e.removeAttribute("href"), e.setAttribute("target", "_blank"), e.setAttribute("rel", "noopener noreferrer");
	return r.body.innerHTML;
}
function Bs(e) {
	return e < 1024 ? e + " B" : e < 1048576 ? (e / 1024).toFixed(1) + " KB" : (e / 1024 / 1024).toFixed(1) + " MB";
}
function Vs(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function Hs(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// plugins/artifacts/Artifacts.svelte
var Us = /* @__PURE__ */ U("<div class=\"card err svelte-pa0w9g\"> </div>"), Ws = /* @__PURE__ */ U("<p class=\"faint\">Loading…</p>"), Gs = /* @__PURE__ */ U("<div class=\"card empty svelte-pa0w9g\" data-testid=\"artifacts-empty\"><div class=\"big svelte-pa0w9g\">Nothing to read.</div> <div class=\"muted\"> </div></div>"), Ks = /* @__PURE__ */ U("<button><div class=\"title svelte-pa0w9g\"> </div> <div class=\"faint\"><!> <!> <!></div></button>"), qs = /* @__PURE__ */ U("<div class=\"err svelte-pa0w9g\"> </div>"), Js = /* @__PURE__ */ U("<span class=\"pill ok\">read</span>"), Ys = /* @__PURE__ */ U("<button data-testid=\"artifact-ack\">Ack</button>"), Xs = /* @__PURE__ */ U("<a> </a>"), Zs = /* @__PURE__ */ U("<span> </span>"), Qs = /* @__PURE__ */ U("<span class=\"mono\"> </span><span class=\"faint\">pruned after retention</span>", 1), $s = /* @__PURE__ */ U("<a class=\"mono\"> </a>"), ec = /* @__PURE__ */ U("<div></div>"), tc = /* @__PURE__ */ U("<img class=\"svelte-pa0w9g\"/>"), nc = /* @__PURE__ */ U("<pre class=\"svelte-pa0w9g\"> </pre>"), rc = /* @__PURE__ */ U("<div class=\"file svelte-pa0w9g\"><div class=\"row svelte-pa0w9g\"><!> <span class=\"faint\"> </span></div> <!></div>"), ic = /* @__PURE__ */ U("<h3 class=\"svelte-pa0w9g\">Attachments</h3> <!>", 1), ac = /* @__PURE__ */ U("<div class=\"comment svelte-pa0w9g\"><div class=\"faint\"><span class=\"mono\"> </span> </div><pre class=\"svelte-pa0w9g\"> </pre></div>"), oc = /* @__PURE__ */ U("<h3 class=\"svelte-pa0w9g\">Comments</h3> <!>", 1), sc = /* @__PURE__ */ U("<header class=\"svelte-pa0w9g\"><h2 class=\"svelte-pa0w9g\"> </h2> <!></header> <div class=\"row faint svelte-pa0w9g\"><span class=\"mono\"> </span> <span> </span> <!> <!> <!> <!></div> <!> <div class=\"md svelte-pa0w9g\" data-testid=\"artifact-body\"></div> <!> <!>", 1), cc = /* @__PURE__ */ U("<div class=\"split svelte-pa0w9g\"><div class=\"card scroll list svelte-pa0w9g\" data-testid=\"artifact-list\"></div> <article class=\"card pane svelte-pa0w9g\" data-testid=\"artifact-pane\"><!></article></div>"), lc = /* @__PURE__ */ U("<h1 class=\"svelte-pa0w9g\">Artifacts</h1> <p class=\"muted sub svelte-pa0w9g\">Reports and files agents delivered. <span class=\"faint\">j/k move · a ack and next · o open source</span></p> <div class=\"filters svelte-pa0w9g\"><input placeholder=\"trigger\" class=\"svelte-pa0w9g\"/> <input placeholder=\"pool\" class=\"svelte-pa0w9g\"/> <label><input type=\"checkbox\" class=\"svelte-pa0w9g\"/> show read</label></div> <!>", 1);
function uc(e, t) {
	Re(t, !0);
	let n = di(() => t.api), r = (e) => t.api.go(e), i = ui(() => t.api, () => {
		let e = new URLSearchParams();
		return n.read || e.set("unread", "1"), n.trigger && e.set("trigger", n.trigger), n.pool && e.set("pool", n.pool), "api/artifacts?" + e.toString();
	}), a = /* @__PURE__ */ at(() => i.data ?? []), o = /* @__PURE__ */ at(() => n.artifact || H(a)[0]?.ticket.id || ""), s = ui(() => t.api, () => H(o) ? "api/artifacts/" + encodeURIComponent(H(o)) : "api/artifacts?unread=1"), c = /* @__PURE__ */ at(() => H(o) && s.data?.ticket?.id === H(o) ? s.data : null), l = /* @__PURE__ */ Pt("");
	function u(e) {
		let t = H(a).findIndex((e) => e.ticket.id === H(o)), n = H(a)[Math.min(Math.max(t + e, 0), H(a).length - 1)];
		n && r({ artifact: n.ticket.id });
	}
	async function d() {
		if (!H(c) || H(c).read) return;
		let e = H(a).findIndex((e) => e.ticket.id === H(o)), i = H(a)[e + 1] ?? H(a)[e - 1];
		It(l, "");
		try {
			await t.api.post("api/artifacts/" + encodeURIComponent(H(o)) + "/ack", {}), r({ artifact: n.read ? H(o) : i && i.ticket.id !== H(o) ? i.ticket.id : "" });
		} catch (e) {
			It(l, fi(e), !0);
		}
	}
	function f() {
		H(c)?.artifact.ticket && r({ ticket: H(c).artifact.ticket });
	}
	function p(e) {
		let t = e.target;
		n.ticket || e.metaKey || e.ctrlKey || e.altKey || t.closest("input, textarea, select") || (e.key === "j" ? u(1) : e.key === "k" ? u(-1) : e.key === "a" ? d() : e.key === "o" && f());
	}
	let m = (e) => !e.pruned && /^image\/(png|jpeg|gif|webp)$/.test(e.type || ""), h = (e) => !e.pruned && e.size <= 262144 && (/^text\//.test(e.type || "") || e.type === "application/json" || !e.type), g = (e) => t.api.url("api/artifacts/" + encodeURIComponent(H(o)) + "/files/" + encodeURIComponent(e));
	async function _(e) {
		let t = await fetch(g(e));
		if (!t.ok) throw Error(t.status + " " + t.statusText);
		return t.text();
	}
	let v = (e) => e.artifact.trigger || e.artifact.pool || "", y = (e) => ({
		type: (e.type || "").split(";")[0],
		ext: (e.name.match(/\.([^.]+)$/)?.[1] || "").toLowerCase()
	}), b = (e) => !e.pruned && t.api.has("artifact", y(e));
	var x = lc();
	lr("keydown", Wt, p);
	var S = R(Qt(x), 4), C = Zt(S);
	ni(C);
	var w = R(C, 2);
	ni(w);
	var ee = R(w, 2), te = Zt(ee);
	ni(te), Se(), A(ee), A(S);
	var T = R(S, 2), ne = (e) => {
		var t = Us(), n = L(t, !0);
		z(() => G(n, i.error)), W(e, t);
	}, re = (e) => {
		W(e, Ws());
	}, ie = (e) => {
		var t = Gs(), r = L(R(Zt(t), 2));
		A(t), z(() => G(r, `No ${n.read ? "" : "unread "}artifacts.`)), W(e, t);
	}, ae = (e) => {
		var n = cc(), i = Zt(n);
		Vr(i, 21, () => H(a), (e) => e.ticket.id, (e, t) => {
			var n = Ks();
			let i;
			var a = Zt(n), s = L(a, !0), c = R(a, 2), l = Zt(c), u = (e) => {
				var n = yr();
				z((e) => G(n, `${e ?? ""} ·`), [() => v(H(t))]), W(e, n);
			}, d = /* @__PURE__ */ at(() => v(H(t)));
			Lr(l, (e) => {
				H(d) && e(u);
			});
			var f = R(l), p = R(f), m = (e) => {
				var n = yr();
				z(() => G(n, `· ${H(t).artifact.files.length ?? ""} file${H(t).artifact.files.length > 1 ? "s" : ""}`)), W(e, n);
			};
			Lr(p, (e) => {
				H(t).artifact.files?.length && e(m);
			});
			var h = R(p, 2), g = (e) => {
				W(e, yr("· read"));
			};
			Lr(h, (e) => {
				H(t).read && e(g);
			}), A(c), A(n), z((e) => {
				i = Zr(n, 1, "item svelte-pa0w9g", null, i, {
					active: H(t).ticket.id === H(o),
					read: H(t).read
				}), G(s, H(t).ticket.title), G(f, `${e ?? ""} `);
			}, [() => Vs(H(t).ticket.created_at)]), ur("click", n, () => r({ artifact: H(t).ticket.id })), W(e, n);
		}), A(i);
		var u = R(i, 2), f = Zt(u), p = (e) => {
			var t = qs(), n = L(t, !0);
			z(() => G(n, s.error)), W(e, t);
		}, x = (e) => {
			let n = /* @__PURE__ */ at(() => H(c).artifact);
			var i = sc(), a = Qt(i), o = Zt(a), s = L(o, !0), u = R(o, 2), f = (e) => {
				W(e, Js());
			}, p = (e) => {
				var t = Ys();
				ur("click", t, d), W(e, t);
			};
			Lr(u, (e) => {
				H(c).read ? e(f) : e(p, -1);
			}), A(a);
			var v = R(a, 2), x = Zt(v), S = L(x, !0), C = R(x, 2), w = L(C, !0), ee = R(C, 2), te = (e) => {
				var t = Xs(), i = L(t);
				z(() => {
					ai(t, "href", "?ticket=" + H(n).ticket), G(i, `source ${H(n).ticket ?? ""}`);
				}), ur("click", t, (e) => {
					e.preventDefault(), r({ ticket: H(n).ticket });
				}), W(e, t);
			};
			Lr(ee, (e) => {
				H(n).ticket && e(te);
			});
			var T = R(ee, 2), ne = (e) => {
				var t = Xs(), i = L(t);
				z(() => {
					ai(t, "href", "?ticket=" + H(n).run), G(i, `run ${H(n).run ?? ""}`);
				}), ur("click", t, (e) => {
					e.preventDefault(), r({ ticket: H(n).run });
				}), W(e, t);
			};
			Lr(T, (e) => {
				H(n).run && e(ne);
			});
			var re = R(T, 2), ie = (e) => {
				var t = Xs(), i = L(t);
				z((e) => {
					ai(t, "href", e), G(i, `trigger ${H(n).trigger ?? ""}`);
				}, [() => "?view=artifacts&read=1&trigger=" + encodeURIComponent(H(n).trigger)]), ur("click", t, (e) => {
					e.preventDefault(), r({
						trigger: H(n).trigger,
						read: "1",
						artifact: ""
					});
				}), W(e, t);
			};
			Lr(re, (e) => {
				H(n).trigger && e(ie);
			});
			var ae = R(re, 2), oe = (e) => {
				var t = Zs(), r = L(t);
				z(() => G(r, `pool ${H(n).pool ?? ""}`)), W(e, t);
			};
			Lr(ae, (e) => {
				H(n).pool && e(oe);
			}), A(v);
			var se = R(v, 2), ce = (e) => {
				var t = qs(), n = L(t, !0);
				z(() => G(n, H(l))), W(e, t);
			};
			Lr(se, (e) => {
				H(l) && e(ce);
			});
			var le = R(se, 2);
			qr(le, () => zs(H(c).ticket.body), !0), A(le);
			var ue = R(le, 2), de = (e) => {
				var r = ic();
				Vr(R(Qt(r), 2), 17, () => H(n).files, (e) => e.name, (e, n) => {
					var r = rc(), i = Zt(r), a = Zt(i), o = (e) => {
						var t = Qs(), r = L(Qt(t), !0);
						Se(), z(() => G(r, H(n).name)), W(e, t);
					}, s = (e) => {
						var t = $s(), r = L(t, !0);
						z((e) => {
							ai(t, "href", e), ai(t, "download", H(n).name), G(r, H(n).name);
						}, [() => g(H(n).name)]), W(e, t);
					};
					Lr(a, (e) => {
						H(n).pruned ? e(o) : e(s, -1);
					});
					var l = L(R(a, 2), !0);
					A(i);
					var u = R(i, 2), d = (e) => {
						var r = ec();
						Jr(r, (e, t) => pi?.(e, t), () => ({
							api: t.api,
							slot: "artifact",
							target: y(H(n)),
							context: {
								artifact: H(c),
								file: {
									...H(n),
									url: g(H(n).name)
								}
							}
						})), W(e, r);
					}, f = /* @__PURE__ */ at(() => b(H(n))), p = (e) => {
						var t = tc();
						z((e) => {
							ai(t, "src", e), ai(t, "alt", H(n).name);
						}, [() => g(H(n).name)]), W(e, t);
					}, v = /* @__PURE__ */ at(() => m(H(n))), x = (e) => {
						var t = br();
						Ir(Qt(t), () => _(H(n).name), null, (e, t) => {
							var n = nc(), r = L(n, !0);
							z(() => G(r, H(t))), W(e, n);
						}, (e, t) => {
							var n = qs(), r = L(n, !0);
							z(() => G(r, H(t).message)), W(e, n);
						}), W(e, t);
					}, S = /* @__PURE__ */ at(() => h(H(n)));
					Lr(u, (e) => {
						H(f) ? e(d) : H(v) ? e(p, 1) : H(S) && e(x, 2);
					}), A(r), z((e) => G(l, e), [() => Bs(H(n).size)]), W(e, r);
				}), W(e, r);
			};
			Lr(ue, (e) => {
				H(n).files?.length && e(de);
			});
			var fe = R(ue, 2), pe = (e) => {
				var t = oc();
				Vr(R(Qt(t), 2), 17, () => H(c).comments, (e) => e.id, (e, t) => {
					var n = ac(), r = Zt(n), i = Zt(r), a = L(i, !0), o = R(i);
					A(r);
					var s = L(R(r), !0);
					A(n), z((e) => {
						G(a, H(t).author), G(o, ` · ${e ?? ""}`), G(s, H(t).body);
					}, [() => Vs(H(t).created_at)]), W(e, n);
				}), W(e, t);
			};
			Lr(fe, (e) => {
				H(c).comments?.length && e(pe);
			}), z((e) => {
				G(s, H(c).ticket.title), G(S, H(c).ticket.id), G(w, e);
			}, [() => Vs(H(c).ticket.created_at)]), W(e, i);
		}, S = (e) => {
			W(e, Ws());
		};
		Lr(f, (e) => {
			s.error && H(o) ? e(p) : H(c) ? e(x, 1) : e(S, -1);
		}), A(u), A(n), W(e, n);
	};
	Lr(T, (e) => {
		i.error ? e(ne) : i.loading && !i.data ? e(re, 1) : H(a).length === 0 ? e(ie, 2) : e(ae, -1);
	}), z(() => {
		ri(C, n.trigger), ri(w, n.pool), ii(te, !!n.read);
	}), ur("change", C, (e) => r({
		trigger: e.currentTarget.value,
		artifact: ""
	})), ur("change", w, (e) => r({
		pool: e.currentTarget.value,
		artifact: ""
	})), ur("change", te, (e) => r({
		read: e.currentTarget.checked ? "1" : "",
		artifact: ""
	})), W(e, x), ze();
}
// plugins/artifacts/index.js
dr(["change", "click"]), li("desk-artifacts-view", uc);
