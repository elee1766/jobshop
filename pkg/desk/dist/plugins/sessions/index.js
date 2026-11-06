// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = Array.isArray, i = Array.prototype.indexOf, a = Array.prototype.includes, o = Array.from, s = Object.defineProperty, c = Object.getOwnPropertyDescriptor, l = Object.getOwnPropertyDescriptors, u = Object.prototype, d = Array.prototype, f = Object.getPrototypeOf, p = Object.isExtensible, m = () => {};
function h(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function g() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, te = 65536, ne = 1 << 19, re = 1 << 20, ie = 1 << 25, ae = 1 << 21, oe = 1 << 22, se = 1 << 23, ce = Symbol("$state"), le = Symbol("component"), ue = Symbol(""), de = Symbol("attributes"), fe = Symbol("class"), pe = Symbol("style"), me = Symbol("text"), he = new class extends Error {
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
var C = !1;
function be(e) {
	C = e;
}
var w;
function T(t) {
	if (t === null) throw ve(), e;
	return w = t;
}
function xe() {
	return T(/* @__PURE__ */ Wt(w));
}
function E(t) {
	if (C) {
		if (/* @__PURE__ */ Wt(w) !== null) throw ve(), e;
		w = t;
	}
}
function Se(e = 1) {
	if (C) {
		for (var t = e, n = w; t--;) n = /* @__PURE__ */ Wt(n);
		w = n;
	}
}
function Ce(e = !0) {
	for (var t = 0, n = w;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Wt(n);
		e && n.remove(), n = i;
	}
}
function we(t) {
	if (!t || t.nodeType !== 8) throw ve(), e;
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
function Pe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Le() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var D = null;
function Re(e) {
	D = e;
}
function ze(e, t = !1, n) {
	D = {
		p: D,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: W,
		l: null
	};
}
function Be(e) {
	var t = D, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) rn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, D = t.p, Ve(e);
}
function Ve(e = {}) {
	return s(e, le, { value: !0 }), e;
}
function He() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ue = [];
function We() {
	var e = Ue;
	Ue = [], h(e);
}
function Ge(e) {
	if (Ue.length === 0 && !pt) {
		var t = Ue;
		queueMicrotask(() => {
			t === Ue && We();
		});
	}
	Ue.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ke = ~(v | y | _);
function O(e, t) {
	e.f = e.f & Ke | t;
}
function qe(e) {
	e.f & 512 || e.deps === null ? O(e, _) : O(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Je(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), O(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Ye(e) {
	var t = V, n = W;
	U(null), G(null);
	try {
		return e();
	} finally {
		U(t), G(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Xe(e, t, n, r) {
	let i = He() ? et : it;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = W, c = Ze(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				R(e, s);
			}
			Qe();
		}
	}
	var d = $e();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ nt(e))).then(u).catch((e) => R(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Qe());
	}) : f();
}
function Ze() {
	var e = W, t = V, n = D, r = k;
	return function(i = !0) {
		G(e), U(t), Re(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Qe(e = !0) {
	G(null), U(null), Re(null), e && k?.deactivate();
}
function $e() {
	var e = W, t = e.b, n = k, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function et(e) {
	var n = 2 | v;
	return W !== null && (W.f |= ne), {
		ctx: D,
		deps: null,
		effects: null,
		equals: Te,
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
var tt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function nt(e, n, r) {
	let i = W;
	i === null && Oe();
	var a = void 0, o = Ot(t), s = !V, c = /* @__PURE__ */ new Set();
	return on(() => {
		var t = W, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== he && n.reject(e);
			}).finally(Qe);
		} catch (e) {
			n.reject(e), Qe();
		}
		var r = k;
		if (s) {
			if (t.f & 32768) var l = $e();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(tt);
			else for (let e of c.values()) e.reject(tt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== tt && (r.activate(), t ? (o.f |= se, Mt(o, t)) : (o.f & 8388608 && (o.f ^= se), Mt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), tn(() => {
		for (let e of c) e.reject(tt);
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
function rt(e) {
	let t = /* @__PURE__ */ et(e);
	return Cn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function it(e) {
	let t = /* @__PURE__ */ et(e);
	return t.equals = De, t;
}
function at(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) B(t[n]);
	}
}
function ot(e) {
	var n, r = W, i = e.parent;
	if (!xn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	G(i);
	try {
		at(e), n = Mn(e);
	} finally {
		G(r);
	}
	return n;
}
function st(e) {
	var t = ot(e);
	!e.equals(t) && (e.wv = kn(), (!k?.is_fork || e.deps === null) && (k === null ? e.v = t : (k.capture(e, t, !0), dt?.capture(e, t, !0)), e.deps === null)) ? O(e, _) : xn || (A === null ? qe(e) : (en() || k?.is_fork) && A.set(e, t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Ye(() => {
		t.ac.abort(he), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Fn(t, 0), dn(t));
}
function lt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && In(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ut = null, k = null, dt = null, A = null, ft = null, pt = !1, mt = !1, ht = null, gt = null, _t = 0, vt = 1, yt = class e {
	id = vt++;
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
		ut === null ? ut = this : (ut.next = this, this.prev = ut), ut = this;
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
			for (var r of n.d) O(r, v), t(r);
			for (r of n.m) O(r, y), t(r);
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
					t.f ^= _;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) O(e, t), this.schedule(e);
		this.apply();
		for (var t = ht = [], n = [], r = gt = []; this.#a.length > 0;) {
			_t++ > 1e3 && (this.#b(), bt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Tt(e), this.#d() || this.discard(), t;
			}
		}
		if (k = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (ht = null, gt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) wt(e, t);
			r.length > 0 && k.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), dt = this, St(n), St(t), dt = null, this.#i?.resolve();
			var o = k;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (j.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : An(r) && (i & 16 && this.add_dirty_reaction(r, y), In(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), O(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), k = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Je(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), A?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		k = this;
	}
	deactivate() {
		k = null, A = null;
	}
	flush() {
		try {
			mt = !0, k = this, this.#m();
		} finally {
			_t = 0, ft = null, ht = null, gt = null, mt = !1, k = null, A = null, j.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(tt);
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
		this.#u || (this.#u = !0, Ge(() => {
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
		return (this.#i ??= g()).promise;
	}
	static ensure() {
		if (k === null) {
			let t = k = new e();
			!mt && Ge(() => {
				t.started || t.flush();
			});
		}
		return k;
	}
	apply() {
		A = null;
	}
	schedule(e) {
		ft = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ut = e : t.prev = e, this.linked = !1;
		}
	}
};
function bt() {
	try {
		Ne();
	} catch (e) {
		R(e, ft);
	}
}
var xt = null;
function St(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && An(r) && (xt = /* @__PURE__ */ new Set(), In(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && mn(r), xt?.size > 0)) {
				j.clear();
				for (let e of xt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) xt.has(n) && (xt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || In(n);
					}
				}
				xt.clear();
			}
		}
		xt = null;
	}
}
function Ct(e) {
	k.schedule(e);
}
function wt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), O(e, _);
		for (var n = e.first; n !== null;) wt(n, t), n = n.next;
	}
}
function Tt(e) {
	O(e, _);
	for (var t = e.first; t !== null;) Tt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Et = /* @__PURE__ */ new Set(), j = /* @__PURE__ */ new Map(), Dt = !1;
function Ot(e, t) {
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
function M(e, t) {
	let n = Ot(e, t);
	return Cn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function kt(e, t = !1, n = !0) {
	let r = Ot(e);
	return t || (r.equals = De), r;
}
function N(e, t, n = !1) {
	return V !== null && (!H || V.f & 131072) && He() && V.f & 4325394 && (K === null || !K.has(e)) && Ie(), Mt(e, n ? Lt(t) : t, gt);
}
var At = null, jt = 0;
function Mt(e, t, n = null) {
	if (!e.equals(t)) {
		xn ? j.set(e, t) : j.has(e) || j.set(e, e.v);
		var r = yt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ot(t), A === null && qe(t);
		}
		e.wv = kn(), At = null, jt = 0, It(e, v, n), At = null, He() && W !== null && W.f & 1024 && !(W.f & 96) && (Y === null ? wn([e]) : Y.push(e)), !r.is_fork && Et.size > 0 && !Dt && Nt();
	}
	return t;
}
function Nt() {
	Dt = !1;
	for (let e of Et) {
		e.f & 1024 && O(e, y);
		let t;
		try {
			t = An(e);
		} catch {
			t = !0;
		}
		t && In(e);
	}
	Et.clear();
}
function Pt(e, t = 1) {
	var n = X(e), r = t === 1 ? n++ : n--;
	return N(e, n), r;
}
function Ft(e) {
	N(e, e.v + 1);
}
function It(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = He(), a = r.length;
		if (jt += a, jt > 1e5 && At === null && (At = /* @__PURE__ */ new Set()), At !== null) {
			if (At.has(e)) return;
			At.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & v) === 0;
				if (l && O(s, t), c & 131072) Et.add(s);
				else if (c & 2) {
					var u = s;
					A?.delete(u), It(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && xt !== null && xt.add(d), n === null ? Ct(d) : n.push(d);
				}
			}
		}
	}
}
function Lt(e) {
	if (typeof e != "object" || !e || ce in e || le in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ M(0), s = null, l = Dn, p = (e) => {
		if (Dn === l) return e();
		var t = V, n = Dn;
		U(null), On(l);
		var r = e();
		return U(t), On(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ M(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Pe();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ M(n.value, s);
				return i.set(t, e), e;
			}) : N(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ M(t, s));
					i.set(n, e), Ft(o);
				}
			} else N(r, t), Ft(o);
			return !0;
		},
		get(n, r, a) {
			if (r === ce) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ M(Lt(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = X(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = X(a);
				if (o === t) return;
				if (r && "value" in r) r.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return r;
		},
		has(e, n) {
			if (n === ce) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (W !== null || V !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ M(a ? Lt(e[n]) : t, s)), i.set(n, r)), X(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ M(t, s)), i.set(f + "", m)) : N(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ M(void 0, s)), N(u, Lt(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => Lt(r));
				N(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && N(_, v + 1);
				}
				Ft(o);
			}
			return !0;
		},
		ownKeys(e) {
			X(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Fe();
		}
	});
}
var Rt, zt, Bt, Vt;
function Ht() {
	if (Rt === void 0) {
		Rt = window, zt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Bt = c(t, "firstChild").get, Vt = c(t, "nextSibling").get, p(e) && (e[fe] = void 0, e[de] = null, e[pe] = void 0, e.__e = void 0), p(n) && (n[me] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ut(e) {
	return Bt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Wt(e) {
	return Vt.call(e);
}
function F(e, t) {
	if (!C) return /* @__PURE__ */ Ut(e);
	var n = /* @__PURE__ */ Ut(w);
	if (n === null) n = w.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), T(r), r;
	}
	return t && Yt(n), T(n), n;
}
function Gt(e, t = !1) {
	if (!C) {
		var n = /* @__PURE__ */ Ut(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Wt(n) : n;
	}
	if (t) {
		if (w?.nodeType !== 3) {
			var r = P();
			return w?.before(r), T(r), r;
		}
		Yt(w);
	}
	return w;
}
function I(e, t = !1) {
	if (!C) return /* @__PURE__ */ Ut(e);
	var n = F(e, t);
	return E(e), n;
}
function L(e, t = 1, n = !1) {
	let r = C ? w : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Wt(r);
	if (!C) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), T(a), a;
		}
		Yt(r);
	}
	return T(r), r;
}
function Kt(e) {
	e.textContent = "";
}
function qt() {
	return !1;
}
function Jt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Yt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Xt(e) {
	var t = W;
	if (t === null) return V.f |= se, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	R(e, t);
}
function R(t, n) {
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
function Zt(e) {
	W === null && (V === null && Me(e), je()), xn && Ae(e);
}
function Qt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function $t(e, t) {
	var n = W;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: D,
		deps: null,
		nodes: null,
		f: e | v | 512,
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
	k?.register_created_effect(r);
	var i = r;
	if (e & 4) ht === null ? yt.ensure().schedule(r) : ht.push(r);
	else if (t !== null) {
		try {
			In(r);
		} catch (e) {
			throw B(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= te));
	}
	if (i !== null && (i.parent = n, n !== null && Qt(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function en() {
	return V !== null && !H;
}
function tn(e) {
	let t = $t(8, null);
	return O(t, _), t.teardown = e, t;
}
function nn(e) {
	Zt("$effect");
	var t = W.f;
	if (!V && t & 32 && D !== null && !D.i) {
		var n = D;
		(n.e ??= []).push(e);
	} else return rn(e);
}
function rn(e) {
	return $t(4 | re, e);
}
function an(e) {
	yt.ensure();
	let t = $t(64 | ne, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? hn(t, () => {
			B(t), n(void 0);
		}) : (B(t), n(void 0));
	});
}
function on(e) {
	return $t(oe | ne, e);
}
function sn(e, t = 0) {
	return $t(8 | t, e);
}
function cn(e, t = [], n = [], r = []) {
	Xe(r, t, n, (t) => {
		$t(8, () => {
			e(...t.map(X));
		});
	});
}
function ln(e, t = 0) {
	return $t(16 | t, e);
}
function z(e) {
	return $t(32 | ne, e);
}
function un(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = xn, r = V;
		Sn(!0), U(null);
		try {
			t.call(null);
		} catch (t) {
			R(t, e.parent);
		} finally {
			Sn(n), U(r);
		}
	}
}
function dn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Ye(() => {
			e.abort(he);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : B(n, t), n = r;
	}
}
function fn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || B(t), t = n;
	}
}
function B(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (pn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, dn(e, t && !n), Fn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	un(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && mn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function pn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Wt(e);
		e.remove(), e = n;
	}
}
function mn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function hn(e, t, n = !0) {
	var r = [];
	e.f |= 256, gn(e, r, !0);
	var i = () => {
		n && B(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function gn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				gn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function _n(e) {
	e.f &= -257, vn(e, !0);
}
function vn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (O(e, v), yt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			vn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function yn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Wt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var bn = null, xn = !1;
function Sn(e) {
	xn = e;
}
var V = null, H = !1;
function U(e) {
	V = e;
}
var W = null;
function G(e) {
	W = e;
}
var K = null;
function Cn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (K ??= /* @__PURE__ */ new Set()).add(e);
}
var q = null, J = 0, Y = null;
function wn(e) {
	Y = e;
}
var Tn = 1, En = 0, Dn = En;
function On(e) {
	Dn = e;
}
function kn() {
	return ++Tn;
}
function An(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (An(a) && st(a), a.wv > e.wv) return !0;
		}
		t & 512 && A === null && O(e, _);
	}
	return !1;
}
function jn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(K !== null && K.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? jn(a, t, !1) : t === a && (n ? O(a, v) : a.f & 1024 && O(a, y), Ct(a));
	}
}
function Mn(e) {
	var t = q, n = J, r = Y, i = V, a = K, o = D, s = H, c = Dn, l = e.f;
	q = null, J = 0, Y = null, V = l & 96 ? null : e, K = null, Re(e.ctx), H = !1, Dn = ++En, e.ac !== null && (Ye(() => {
		e.ac.abort(he);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Nn(e);
		if (He() && Y !== null && !H && f !== null && !(e.f & 6146)) for (var p = 0; p < Y.length; p++) jn(Y[p], e);
		if (i !== null && i !== e) {
			if (En++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = En;
			if (t !== null) for (let e of t) e.rv = En;
			Y !== null && (r === null ? r = Y : r.push(...Y));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return Nn(e), Xt(t);
	} finally {
		e.f ^= ae, q = t, J = n, Y = r, V = i, K = a, Re(o), H = s, Dn = c;
	}
}
function Nn(e) {
	var t = e.deps, n = k?.is_fork;
	if (q !== null) {
		var r;
		if (n || Fn(e, J), t !== null && J > 0) for (t.length = J + q.length, r = 0; r < q.length; r++) t[J + r] = q[r];
		else e.deps = t = q;
		if (en() && e.f & 512) for (r = J; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && J < t.length && (Fn(e, J), t.length = J);
	return t;
}
function Pn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (q === null || !a.call(q, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && qe(c), c.ac !== null && Ye(() => {
			c.ac.abort(he), c.ac = null, O(c, v);
		}), ct(c), Fn(c, 0);
	}
}
function Fn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Pn(e, n[r]);
}
function In(e) {
	var t = e.f;
	if (!(t & 16384)) {
		O(e, _);
		var n = W;
		W = e;
		try {
			t & 16777232 ? fn(e) : dn(e), un(e);
			var r = Mn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Tn;
		} finally {
			W = n;
		}
	}
}
function X(e) {
	var t = !!(e.f & 2);
	if (bn?.add(e), V !== null && !H && !(W !== null && W.f & 16384) && (K === null || !K.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < En && (e.rv = En, q === null && n !== null && n[J] === e ? J++ : q === null ? q = [e] : q.push(e));
		else {
			V.deps ??= [], a.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : a.call(r, V) || r.push(V);
		}
	}
	if (xn && j.has(e)) return j.get(e);
	if (t) {
		var i = e;
		if (xn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Rn(i)) && (o = ot(i)), j.set(i, o), o;
		}
		var s = !(i.f & 512) && !H && V !== null && !!(V.f & 512), c = (i.f & S) === 0;
		An(i) && (s && (i.f |= 512), st(i)), s && !c && (lt(i), Ln(i));
	}
	if (A?.has(e)) return A.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Ln(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (lt(n), Ln(n));
	}
}
function Rn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (j.has(t) || t.f & 2 && Rn(t)) return !0;
	return !1;
}
function zn(e) {
	var t = H;
	try {
		return H = !0, e();
	} finally {
		H = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Bn = Symbol("events"), Vn = /* @__PURE__ */ new Set(), Hn = /* @__PURE__ */ new Set();
function Un(e, t, n) {
	(t[Bn] ??= {})[e] = n;
}
function Wn(e) {
	for (var t = 0; t < e.length; t++) Vn.add(e[t]);
	for (var n of Hn) n(e);
}
var Gn = null, Kn = !1;
function qn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Gn = e, Kn || (Kn = !0, setTimeout(() => {
		Kn = !1, Gn = null;
	}));
	var o = 0, c = Gn === e && e[Bn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Bn] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (o = l);
	}
	if (a = i[o] || e.target, a !== t) {
		s(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = V, f = W;
		U(null), G(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Bn]?.[r];
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
			e[Bn] = t, delete e.currentTarget, U(d), G(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Jn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Yn(e) {
	return Jn?.createHTML(e) ?? e;
}
function Xn(e) {
	var t = Jt("template");
	return t.innerHTML = Yn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Zn(e, t) {
	var n = W;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Z(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (C) return Zn(w, null), w;
		i === void 0 && (i = Xn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Ut(i)));
		var t = r || zt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Ut(t), s = t.lastChild;
			Zn(o, s);
		} else Zn(t, t);
		return t;
	};
}
function Qn() {
	if (C) return Zn(w, null), w;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = P();
	return e.append(t, n), Zn(t, n), e;
}
function Q(e, t) {
	if (C) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = w), xe();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var $n = ["touchstart", "touchmove"];
function er(e) {
	return $n.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function tr(e) {
	let t = 0, n = Ot(0), r;
	return () => {
		en() && (X(n), sn(() => (t === 0 && (r = zn(() => e(() => Ft(n)))), t += 1, () => {
			Ge(() => {
				--t, t === 0 && (r?.(), r = void 0, Ft(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var nr = te | ne;
function rr(e, t, n, r) {
	new ir(e, t, n, r);
}
var ir = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = C ? w : null;
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
	#m = tr(() => (this.#p = Ot(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = ln(() => {
			if (C) {
				let e = we(this.#t);
				xe();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, nr), C && (this.#e = w);
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
		Ge(r), t && (this.#s = z(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Le(), this.#s !== null && hn(this.#s, () => {
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
					R(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = z(() => e(this.#e)), Ge(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = P(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return z(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							R(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(k)) : this.#u === 0 && (this.#e.before(e), this.#c = null, hn(this.#o, () => {
					this.#o = null;
				}), this.#x(k));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = z(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				yn(this.#a, e);
				let t = this.#n.pending;
				this.#o = z(() => t(this.#e));
			} else this.#x(k);
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
		var t = W, n = V, r = D;
		G(this.#i), U(this.#i), Re(this.#i.ctx);
		try {
			return yt.ensure(), e();
		} finally {
			G(t), U(n), Re(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && hn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ge(() => {
			this.#d = !1, this.#p && Mt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), X(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		k?.is_fork ? (this.#a && k.skip_effect(this.#a), this.#o && k.skip_effect(this.#o), this.#s && k.skip_effect(this.#s), k.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (B(this.#a), null), this.#o &&= (B(this.#o), null), this.#s &&= (B(this.#s), null), C && (T(this.#t), Se(), T(Ce()));
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
					return R(e, this.#i.parent), null;
				}
			}));
		};
		Ge(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					R(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => R(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function $(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[me] ??= e.nodeValue) && (e[me] = n, e.nodeValue = `${n}`);
}
function ar(e, t) {
	return sr(e, t);
}
var or = /* @__PURE__ */ new Map();
function sr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Ht();
	var u = void 0, d = an(() => {
		var c = r ?? n.appendChild(P());
		rr(c, { pending: () => {} }, (n) => {
			ze({});
			var r = D;
			if (s && (r.c = s), a && (i.$$events = a), C && Zn(n, null), u = t(n, i) || Ve(), C && (W.nodes.end = w, w === null || w.nodeType !== 8 || w.data !== "]")) throw ve(), e;
			Be();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = er(r);
					for (let e of [n, document]) {
						var a = or.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), or.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, qn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Vn)), Hn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = or.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, qn), t.delete(e), t.size === 0 && or.delete(r)) : t.set(e, i);
			}
			Hn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return cr.set(u, d), u;
}
var cr = /* @__PURE__ */ new WeakMap();
function lr(e, t) {
	let n = cr.get(e);
	return n ? (cr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var ur = class {
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
			if (n) _n(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (_n(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
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
						yn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else B(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), hn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (B(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = k, r = qt();
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
		} else C && (this.anchor = w), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function dr(e, t, n = !1) {
	var r;
	C && (r = w, xe());
	var i = new ur(e), a = n ? te : 0;
	function o(e, t) {
		if (C) {
			var n = we(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Ce();
				T(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	ln(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function fr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		hn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					pr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Kt(d), d.append(u), e.items.clear();
		}
		pr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function pr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ie, yn(a, document.createDocumentFragment())) : B(t[i], n);
	}
}
var mr;
function hr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = C ? T(/* @__PURE__ */ Ut(u)) : u.appendChild(P());
	}
	C && xe();
	var d = null, f = /* @__PURE__ */ it(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = X(f);
			_.fallback = d, _r(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, yr(d, null, c)) : _n(d) : hn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: ln(() => {
			var e = X(f), r = e.length;
			let o = !1;
			C && we(c) === "[!" != (r === 0) && (c = Ce(), T(c), be(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = k, v = qt(), y = 0; y < r; y += 1) {
				C && w.nodeType === 8 && w.data === "]" && (c = w, o = !0, be(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Mt(S.v, b), S.i && Mt(S.i, y), v && _.unskip_effect(S.e)) : (S = vr(l, m ? c : mr ??= P(), b, x, y, a, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = z(() => s(c)) : (d = z(() => s(mr ??= P())), d.f |= ie)), r > u.size && ke("", "", ""), C && r > 0 && T(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && be(!0), X(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, C && (c = w);
}
function gr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function _r(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = gr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (_n(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ie, _ === l) yr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), br(e, d, _), br(e, _, y), yr(_, y, n), d = _, p = [], m = [], l = gr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) yr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					br(e, S.prev, ee.next), br(e, d, S), br(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), yr(_, l, n), br(e, _.prev, _.next), br(e, _, d === null ? e.effect.first : d.next), br(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = gr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = gr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (pr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var te = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || te.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && te.push(l), l = gr(l.next);
		var ne = te.length;
		if (ne > 0) {
			var re = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < ne; v += 1) te[v].nodes?.a?.measure();
				for (v = 0; v < ne; v += 1) te[v].nodes?.a?.fix();
			}
			fr(e, te, re);
		}
	}
	a && Ge(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function vr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Ot(n) : /* @__PURE__ */ kt(n, !1, !1) : null, l = o & 2 ? Ot(i) : null;
	return {
		v: c,
		i: l,
		e: z(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function yr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Wt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function br(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var xr = [..." 	\n\r\f\xA0\v﻿"];
function Sr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || xr.includes(r[o - 1])) && (s === r.length || xr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Cr(e, t, n, r, i, a) {
	var o = e[fe];
	if (C || o !== n || o === void 0) {
		var s = Sr(n, r, a);
		C && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[fe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var wr = Symbol("is custom element"), Tr = Symbol("is html"), Er = ge ? "link" : "LINK";
function Dr(e, t, n, r) {
	var i = Or(e);
	if (!(C && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Er))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ue] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Ar(e).has(t) ? e[t] = n : (!C || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Or(e) {
	return e[de] ??= {
		[wr]: e.nodeName.includes("-"),
		[Tr]: e.namespaceURI === n
	};
}
var kr = /* @__PURE__ */ new Map();
function Ar(e) {
	var t = e.getAttribute("is") || e.nodeName, n = kr.get(t);
	if (n) return n;
	kr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// kit/plugin.svelte.js
function jr(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ M(Lt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return X(this.#e);
		}
		set props(e) {
			N(this.#e, e, !0);
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
			this.style.display = "contents", this.app ||= ar(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && lr(this.app), this.app = null;
		}
	});
}
function Mr(e, t) {
	let n = Lt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Pt(r);
		}
	}), r = /* @__PURE__ */ M(0);
	nn(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Pt(r);
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
	return nn(() => {
		let n = e(), c = t();
		X(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function Nr(e) {
	let t = Lt({});
	return nn(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
var Pr = (e) => String(e?.message || e);
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/format.js
function Fr(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
// plugins/sessions/Sessions.svelte
var Ir = /* @__PURE__ */ Z("<span class=\"sub\"> </span>"), Lr = /* @__PURE__ */ Z("<div class=\"card err\"> </div>"), Rr = /* @__PURE__ */ Z("<div class=\"muted\"> </div>"), zr = /* @__PURE__ */ Z("<span class=\"pill bad gap svelte-1yaxaek\"> </span>"), Br = /* @__PURE__ */ Z("<span class=\"pill warn gap svelte-1yaxaek\" title=\"no new claims\"> </span>"), Vr = /* @__PURE__ */ Z("<a class=\"mono tk svelte-1yaxaek\"> </a>"), Hr = /* @__PURE__ */ Z("<span class=\"pill warn gap svelte-1yaxaek\">override</span>"), Ur = /* @__PURE__ */ Z("<button class=\"svelte-1yaxaek\">Uncordon</button>"), Wr = /* @__PURE__ */ Z("<button class=\"svelte-1yaxaek\">Cordon</button>"), Gr = /* @__PURE__ */ Z("<tr><td class=\"c-id svelte-1yaxaek\"><a class=\"mono sl\"> </a><!><!></td><td class=\"c-st svelte-1yaxaek\"><span> </span><!></td><td class=\"svelte-1yaxaek\"><!></td><td class=\"muted svelte-1yaxaek\"> </td><td class=\"muted mono svelte-1yaxaek\"> <!></td><td class=\"muted c-time svelte-1yaxaek\"> </td><td class=\"c-act svelte-1yaxaek\"><!></td></tr>"), Kr = /* @__PURE__ */ Z("<tr><td colspan=\"7\" class=\"faint\">No live sessions.</td></tr>"), qr = /* @__PURE__ */ Z("<!> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"sessions\"><thead><tr><th>Session</th><th>State</th><th>Ticket</th><th>Pool</th><th>Model</th><th>Last event</th><th></th></tr></thead><tbody></tbody></table></div>", 1), Jr = /* @__PURE__ */ Z("<header class=\"vh\"><h1>Sessions</h1> <!></header> <div class=\"pad\"><!></div>", 1);
function Yr(e, t) {
	ze(t, !0);
	let n = Nr(() => t.api), r = (e) => t.api.go(e), i = () => t.api.closeSession(), a = Mr(() => t.api, () => "api/sessions"), o = /* @__PURE__ */ M("");
	async function s(e, n) {
		e.stopPropagation(), N(o, "");
		try {
			await t.api.post("api/session/cordon", n), a.refresh();
		} catch (e) {
			N(o, Pr(e), !0);
		}
	}
	function c(e) {
		return e.alive ? e.backoff_until ? ["backoff", "warn"] : e.drained ? ["drained", "warn"] : e.current ? ["working", "ok"] : ["idle", ""] : ["dead", "bad"];
	}
	var l = Jr(), u = Gt(l), d = L(F(u), 2), f = (e) => {
		var t = Ir(), n = I(t);
		cn(() => $(n, `${a.data.length ?? ""} live`)), Q(e, t);
	};
	dr(d, (e) => {
		a.data && e(f);
	}), E(u);
	var p = L(u, 2), m = F(p), h = (e) => {
		var t = Lr(), n = I(t, !0);
		cn(() => $(n, a.error)), Q(e, t);
	}, g = (e) => {
		var t = qr(), l = Gt(t), u = (e) => {
			var t = Lr(), n = I(t, !0);
			cn(() => $(n, X(o))), Q(e, t);
		};
		dr(l, (e) => {
			X(o) && e(u);
		});
		var d = L(l, 2), f = F(d), p = L(F(f));
		hr(p, 21, () => a.data, (e) => e.id, (e, t) => {
			let a = /* @__PURE__ */ rt(() => c(X(t)));
			var o = Gr();
			let l;
			var u = F(o), d = F(u), f = I(d, !0), p = L(d), m = (e) => {
				var n = Rr(), r = I(n, !0);
				cn(() => $(r, X(t).display_name)), Q(e, n);
			};
			dr(p, (e) => {
				X(t).display_name && e(m);
			});
			var h = L(p), g = (e) => {
				var n = zr(), r = I(n);
				cn(() => $(r, `${X(t).crashes ?? ""} crashes`)), Q(e, n);
			};
			dr(h, (e) => {
				X(t).crashes && e(g);
			}), E(u);
			var _ = L(u), v = F(_), y = I(v, !0), b = L(v), x = (e) => {
				var n = Br(), r = I(n);
				cn(() => $(r, `cordoned: ${X(t).cordon ?? ""}`)), Q(e, n);
			};
			dr(b, (e) => {
				X(t).cordon && e(x);
			}), E(_);
			var S = L(_), ee = F(S), te = (e) => {
				var n = Vr(), i = I(n, !0);
				cn(() => {
					Dr(n, "href", "?ticket=" + X(t).current), $(i, X(t).current);
				}), Un("click", n, (e) => {
					e.preventDefault(), e.stopPropagation(), r({ ticket: X(t).current });
				}), Q(e, n);
			};
			dr(ee, (e) => {
				X(t).current && e(te);
			}), E(S);
			var ne = L(S), re = I(ne, !0), ie = L(ne), ae = F(ie), oe = L(ae), se = (e) => {
				Q(e, Hr());
			};
			dr(oe, (e) => {
				X(t).override && e(se);
			}), E(ie);
			var ce = L(ie), le = I(ce, !0), ue = L(ce), de = F(ue), fe = (e) => {
				var n = Qn(), r = Gt(n), i = (e) => {
					var n = Ur();
					Un("click", n, (e) => s(e, {
						id: X(t).id,
						uncordon: !0
					})), Q(e, n);
				}, a = (e) => {
					var n = Wr();
					Un("click", n, (e) => s(e, { id: X(t).id })), Q(e, n);
				};
				dr(r, (e) => {
					X(t).cordon ? e(i) : e(a, -1);
				}), Q(e, n);
			};
			dr(de, (e) => {
				X(t).alive && e(fe);
			}), E(ue), E(o), cn((e) => {
				l = Cr(o, 1, "click svelte-1yaxaek", null, l, { sel: n.session === X(t).id }), Dr(d, "href", "?session=" + X(t).id), $(f, X(t).id), Cr(v, 1, `pill ${X(a)[1] ?? ""}`, "svelte-1yaxaek"), $(y, X(a)[0]), $(re, X(t).ticket?.pool ?? ""), $(ae, `${X(t).model ?? "" ?? ""}${X(t).effort ? " · " + X(t).effort : ""}`), $(le, e);
			}, [() => Fr(X(t).seq_time)]), Un("click", o, () => n.session === X(t).id ? i() : r({ session: X(t).id })), Un("click", d, (e) => {
				e.preventDefault(), e.stopPropagation(), r({
					session: X(t).id,
					ticket: ""
				});
			}), Q(e, o);
		}, (e) => {
			Q(e, Kr());
		}), E(p), E(f), E(d), Q(e, t);
	};
	dr(m, (e) => {
		a.error ? e(h) : a.data && e(g, 1);
	}), E(p), Q(e, l), Be();
}
// plugins/sessions/index.js
Wn(["click"]), jr("desk-sessions-view", Yr);
