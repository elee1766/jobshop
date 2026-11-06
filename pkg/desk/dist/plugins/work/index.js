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
	return T(/* @__PURE__ */ Kt(w));
}
function E(t) {
	if (C) {
		if (/* @__PURE__ */ Kt(w) !== null) throw ve(), e;
		w = t;
	}
}
function Se(e = 1) {
	if (C) {
		for (var t = e, n = w; t--;) n = /* @__PURE__ */ Kt(n);
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
		var i = /* @__PURE__ */ Kt(n);
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
// node_modules/svelte/src/internal/shared/clone.js
var Re = [];
function ze(e, t = !1, n = !1) {
	return Be(e, /* @__PURE__ */ new Map(), "", Re, null, n);
}
function Be(e, t, n, i, a = null, o = !1) {
	if (typeof e == "object" && e) {
		var s = t.get(e);
		if (s !== void 0) return s;
		if (e instanceof Map) return new Map(e);
		if (e instanceof Set) return new Set(e);
		if (r(e)) {
			var c = Array(e.length);
			t.set(e, c), a !== null && t.set(a, c);
			for (var l = 0; l < e.length; l += 1) {
				var d = e[l];
				l in e && (c[l] = Be(d, t, n, i, null, o));
			}
			return c;
		}
		if (f(e) === u) {
			c = {}, t.set(e, c), a !== null && t.set(a, c);
			for (var p of Object.keys(e)) c[p] = Be(e[p], t, n, i, null, o);
			return c;
		}
		if (e instanceof Date) return e.getTime(), structuredClone(e);
		if (typeof e.toJSON == "function" && !o) return Be(e.toJSON(), t, n, i, e);
	}
	if (e instanceof EventTarget) return e;
	try {
		return structuredClone(e);
	} catch {
		return e;
	}
}
// node_modules/svelte/src/internal/client/context.js
var D = null;
function Ve(e) {
	D = e;
}
function He(e, t = !1, n) {
	D = {
		p: D,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: G,
		l: null
	};
}
function Ue(e) {
	var t = D, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) on(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, D = t.p, We(e);
}
function We(e = {}) {
	return s(e, le, { value: !0 }), e;
}
function Ge() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ke = [];
function qe() {
	var e = Ke;
	Ke = [], h(e);
}
function Je(e) {
	if (Ke.length === 0 && !gt) {
		var t = Ke;
		queueMicrotask(() => {
			t === Ke && qe();
		});
	}
	Ke.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ye = ~(v | y | _);
function O(e, t) {
	e.f = e.f & Ye | t;
}
function Xe(e) {
	e.f & 512 || e.deps === null ? O(e, _) : O(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ze(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), O(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Qe(e) {
	var t = H, n = G;
	W(null), K(null);
	try {
		return e();
	} finally {
		W(t), K(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function $e(e, t, n, r) {
	let i = Ge() ? rt : st;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = G, c = et(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				z(e, s);
			}
			tt();
		}
	}
	var d = nt();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ at(e))).then(u).catch((e) => z(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), tt());
	}) : f();
}
function et() {
	var e = G, t = H, n = D, r = k;
	return function(i = !0) {
		K(e), W(t), Ve(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function tt(e = !0) {
	K(null), W(null), Ve(null), e && k?.deactivate();
}
function nt() {
	var e = G, t = e.b, n = k, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function rt(e) {
	var n = 2 | v;
	return G !== null && (G.f |= ne), {
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
		parent: G,
		ac: null
	};
}
var it = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function at(e, n, r) {
	let i = G;
	i === null && Oe();
	var a = void 0, o = At(t), s = !H, c = /* @__PURE__ */ new Set();
	return ln(() => {
		var t = G, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== he && n.reject(e);
			}).finally(tt);
		} catch (e) {
			n.reject(e), tt();
		}
		var r = k;
		if (s) {
			if (t.f & 32768) var l = nt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(it);
			else for (let e of c.values()) e.reject(it);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== it && (r.activate(), t ? (o.f |= se, Pt(o, t)) : (o.f & 8388608 && (o.f ^= se), Pt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), rn(() => {
		for (let e of c) e.reject(it);
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
function ot(e) {
	let t = /* @__PURE__ */ rt(e);
	return En(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function st(e) {
	let t = /* @__PURE__ */ rt(e);
	return t.equals = De, t;
}
function ct(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) V(t[n]);
	}
}
function lt(e) {
	var n, r = G, i = e.parent;
	if (!wn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	K(i);
	try {
		ct(e), n = Fn(e);
	} finally {
		K(r);
	}
	return n;
}
function ut(e) {
	var t = lt(e);
	!e.equals(t) && (e.wv = Mn(), (!k?.is_fork || e.deps === null) && (k === null ? e.v = t : (k.capture(e, t, !0), mt?.capture(e, t, !0)), e.deps === null)) ? O(e, _) : wn || (A === null ? Xe(e) : (nn() || k?.is_fork) && A.set(e, t));
}
function dt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Qe(() => {
		t.ac.abort(he), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Rn(t, 0), mn(t));
}
function ft(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && zn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var pt = null, k = null, mt = null, A = null, ht = null, gt = !1, _t = !1, vt = null, yt = null, bt = 0, xt = 1, St = class e {
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
		pt === null ? pt = this : (pt.next = this, this.prev = pt), pt = this;
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
		for (var t = vt = [], n = [], r = yt = []; this.#a.length > 0;) {
			bt++ > 1e3 && (this.#b(), Ct());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Dt(e), this.#d() || this.discard(), t;
			}
		}
		if (k = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (vt = null, yt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Et(e, t);
			r.length > 0 && k.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), mt = this, wt(n), wt(t), mt = null, this.#i?.resolve();
			var o = k;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (M.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : Nn(r) && (i & 16 && this.add_dirty_reaction(r, y), zn(r));
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
		for (var t = 0; t < e.length; t += 1) Ze(e[t], this.#s);
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
			_t = !0, k = this, this.#m();
		} finally {
			bt = 0, ht = null, vt = null, yt = null, _t = !1, k = null, A = null, M.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(it);
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
		this.#u || (this.#u = !0, Je(() => {
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
			!_t && Je(() => {
				t.started || t.flush();
			});
		}
		return k;
	}
	apply() {
		A = null;
	}
	schedule(e) {
		ht = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? pt = e : t.prev = e, this.linked = !1;
		}
	}
};
function Ct() {
	try {
		Ne();
	} catch (e) {
		z(e, ht);
	}
}
var j = null;
function wt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Nn(r) && (j = /* @__PURE__ */ new Set(), zn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && _n(r), j?.size > 0)) {
				M.clear();
				for (let e of j) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) j.has(n) && (j.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || zn(n);
					}
				}
				j.clear();
			}
		}
		j = null;
	}
}
function Tt(e) {
	k.schedule(e);
}
function Et(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), O(e, _);
		for (var n = e.first; n !== null;) Et(n, t), n = n.next;
	}
}
function Dt(e) {
	O(e, _);
	for (var t = e.first; t !== null;) Dt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Ot = /* @__PURE__ */ new Set(), M = /* @__PURE__ */ new Map(), kt = !1;
function At(e, t) {
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
function N(e, t) {
	let n = At(e, t);
	return En(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function jt(e, t = !1, n = !0) {
	let r = At(e);
	return t || (r.equals = De), r;
}
function P(e, t, n = !1) {
	return H !== null && (!U || H.f & 131072) && Ge() && H.f & 4325394 && (q === null || !q.has(e)) && Ie(), Pt(e, n ? zt(t) : t, yt);
}
var Mt = null, Nt = 0;
function Pt(e, t, n = null) {
	if (!e.equals(t)) {
		wn ? M.set(e, t) : M.has(e) || M.set(e, e.v);
		var r = St.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && lt(t), A === null && Xe(t);
		}
		e.wv = Mn(), Mt = null, Nt = 0, Rt(e, v, n), Mt = null, Ge() && G !== null && G.f & 1024 && !(G.f & 96) && (X === null ? Dn([e]) : X.push(e)), !r.is_fork && Ot.size > 0 && !kt && Ft();
	}
	return t;
}
function Ft() {
	kt = !1;
	for (let e of Ot) {
		e.f & 1024 && O(e, y);
		let t;
		try {
			t = Nn(e);
		} catch {
			t = !0;
		}
		t && zn(e);
	}
	Ot.clear();
}
function It(e, t = 1) {
	var n = Z(e), r = t === 1 ? n++ : n--;
	return P(e, n), r;
}
function Lt(e) {
	P(e, e.v + 1);
}
function Rt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ge(), a = r.length;
		if (Nt += a, Nt > 1e5 && Mt === null && (Mt = /* @__PURE__ */ new Set()), Mt !== null) {
			if (Mt.has(e)) return;
			Mt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== G) {
				var l = (c & v) === 0;
				if (l && O(s, t), c & 131072) Ot.add(s);
				else if (c & 2) {
					var u = s;
					A?.delete(u), Rt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && j !== null && j.add(d), n === null ? Tt(d) : n.push(d);
				}
			}
		}
	}
}
function zt(e) {
	if (typeof e != "object" || !e || ce in e || le in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ N(0), s = null, l = An, p = (e) => {
		if (An === l) return e();
		var t = H, n = An;
		W(null), jn(l);
		var r = e();
		return W(t), jn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ N(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Pe();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ N(n.value, s);
				return i.set(t, e), e;
			}) : P(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ N(t, s));
					i.set(n, e), Lt(o);
				}
			} else P(r, t), Lt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === ce) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ N(zt(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = Z(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = Z(a);
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
			return (r !== void 0 || (G !== null || H !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ N(a ? zt(e[n]) : t, s)), i.set(n, r)), Z(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ N(t, s)), i.set(f + "", m)) : P(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ N(void 0, s)), P(u, zt(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => zt(r));
				P(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && P(_, v + 1);
				}
				Lt(o);
			}
			return !0;
		},
		ownKeys(e) {
			Z(o);
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
var Bt, Vt, Ht, Ut;
function Wt() {
	if (Bt === void 0) {
		Bt = window, Vt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Ht = c(t, "firstChild").get, Ut = c(t, "nextSibling").get, p(e) && (e[fe] = void 0, e[de] = null, e[pe] = void 0, e.__e = void 0), p(n) && (n[me] = void 0);
	}
}
function F(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Gt(e) {
	return Ht.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Kt(e) {
	return Ut.call(e);
}
function I(e, t) {
	if (!C) return /* @__PURE__ */ Gt(e);
	var n = /* @__PURE__ */ Gt(w);
	if (n === null) n = w.appendChild(F());
	else if (t && n.nodeType !== 3) {
		var r = F();
		return n?.before(r), T(r), r;
	}
	return t && Zt(n), T(n), n;
}
function qt(e, t = !1) {
	if (!C) {
		var n = /* @__PURE__ */ Gt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Kt(n) : n;
	}
	if (t) {
		if (w?.nodeType !== 3) {
			var r = F();
			return w?.before(r), T(r), r;
		}
		Zt(w);
	}
	return w;
}
function L(e, t = !1) {
	if (!C) return /* @__PURE__ */ Gt(e);
	var n = I(e, t);
	return E(e), n;
}
function R(e, t = 1, n = !1) {
	let r = C ? w : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Kt(r);
	if (!C) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = F();
			return r === null ? i?.after(a) : r.before(a), T(a), a;
		}
		Zt(r);
	}
	return T(r), r;
}
function Jt(e) {
	e.textContent = "";
}
function Yt() {
	return !1;
}
function Xt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Zt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Qt(e) {
	var t = G;
	if (t === null) return H.f |= se, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	z(e, t);
}
function z(t, n) {
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
function $t(e) {
	G === null && (H === null && Me(e), je()), wn && Ae(e);
}
function en(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function tn(e, t) {
	var n = G;
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
	if (e & 4) vt === null ? St.ensure().schedule(r) : vt.push(r);
	else if (t !== null) {
		try {
			zn(r);
		} catch (e) {
			throw V(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= te));
	}
	if (i !== null && (i.parent = n, n !== null && en(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function nn() {
	return H !== null && !U;
}
function rn(e) {
	let t = tn(8, null);
	return O(t, _), t.teardown = e, t;
}
function an(e) {
	$t("$effect");
	var t = G.f;
	if (!H && t & 32 && D !== null && !D.i) {
		var n = D;
		(n.e ??= []).push(e);
	} else return on(e);
}
function on(e) {
	return tn(4 | re, e);
}
function sn(e) {
	St.ensure();
	let t = tn(64 | ne, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? vn(t, () => {
			V(t), n(void 0);
		}) : (V(t), n(void 0));
	});
}
function cn(e) {
	return tn(4, e);
}
function ln(e) {
	return tn(oe | ne, e);
}
function un(e, t = 0) {
	return tn(8 | t, e);
}
function dn(e, t = [], n = [], r = []) {
	$e(r, t, n, (t) => {
		tn(8, () => {
			e(...t.map(Z));
		});
	});
}
function fn(e, t = 0) {
	return tn(16 | t, e);
}
function B(e) {
	return tn(32 | ne, e);
}
function pn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = wn, r = H;
		Tn(!0), W(null);
		try {
			t.call(null);
		} catch (t) {
			z(t, e.parent);
		} finally {
			Tn(n), W(r);
		}
	}
}
function mn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Qe(() => {
			e.abort(he);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : V(n, t), n = r;
	}
}
function hn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || V(t), t = n;
	}
}
function V(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (gn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, mn(e, t && !n), Rn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	pn(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && _n(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function gn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Kt(e);
		e.remove(), e = n;
	}
}
function _n(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function vn(e, t, n = !0) {
	var r = [];
	e.f |= 256, yn(e, r, !0);
	var i = () => {
		n && V(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function yn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				yn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function bn(e) {
	e.f &= -257, xn(e, !0);
}
function xn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (O(e, v), St.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			xn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Sn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Kt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Cn = null, wn = !1;
function Tn(e) {
	wn = e;
}
var H = null, U = !1;
function W(e) {
	H = e;
}
var G = null;
function K(e) {
	G = e;
}
var q = null;
function En(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (q ??= /* @__PURE__ */ new Set()).add(e);
}
var J = null, Y = 0, X = null;
function Dn(e) {
	X = e;
}
var On = 1, kn = 0, An = kn;
function jn(e) {
	An = e;
}
function Mn() {
	return ++On;
}
function Nn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Nn(a) && ut(a), a.wv > e.wv) return !0;
		}
		t & 512 && A === null && O(e, _);
	}
	return !1;
}
function Pn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(q !== null && q.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Pn(a, t, !1) : t === a && (n ? O(a, v) : a.f & 1024 && O(a, y), Tt(a));
	}
}
function Fn(e) {
	var t = J, n = Y, r = X, i = H, a = q, o = D, s = U, c = An, l = e.f;
	J = null, Y = 0, X = null, H = l & 96 ? null : e, q = null, Ve(e.ctx), U = !1, An = ++kn, e.ac !== null && (Qe(() => {
		e.ac.abort(he);
	}), e.ac = null);
	try {
		e.f |= ae;
		var u = e.fn, d = u();
		e.f |= S;
		var f = In(e);
		if (Ge() && X !== null && !U && f !== null && !(e.f & 6146)) for (var p = 0; p < X.length; p++) Pn(X[p], e);
		if (i !== null && i !== e) {
			if (kn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = kn;
			if (t !== null) for (let e of t) e.rv = kn;
			X !== null && (r === null ? r = X : r.push(...X));
		}
		return e.f & 8388608 && (e.f ^= se), d;
	} catch (t) {
		return In(e), Qt(t);
	} finally {
		e.f ^= ae, J = t, Y = n, X = r, H = i, q = a, Ve(o), U = s, An = c;
	}
}
function In(e) {
	var t = e.deps, n = k?.is_fork;
	if (J !== null) {
		var r;
		if (n || Rn(e, Y), t !== null && Y > 0) for (t.length = Y + J.length, r = 0; r < J.length; r++) t[Y + r] = J[r];
		else e.deps = t = J;
		if (nn() && e.f & 512) for (r = Y; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Y < t.length && (Rn(e, Y), t.length = Y);
	return t;
}
function Ln(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (J === null || !a.call(J, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Xe(c), c.ac !== null && Qe(() => {
			c.ac.abort(he), c.ac = null, O(c, v);
		}), dt(c), Rn(c, 0);
	}
}
function Rn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Ln(e, n[r]);
}
function zn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		O(e, _);
		var n = G;
		G = e;
		try {
			t & 16777232 ? hn(e) : mn(e), pn(e);
			var r = Fn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = On;
		} finally {
			G = n;
		}
	}
}
function Z(e) {
	var t = !!(e.f & 2);
	if (Cn?.add(e), H !== null && !U && !(G !== null && G.f & 16384) && (q === null || !q.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < kn && (e.rv = kn, J === null && n !== null && n[Y] === e ? Y++ : J === null ? J = [e] : J.push(e));
		else {
			H.deps ??= [], a.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : a.call(r, H) || r.push(H);
		}
	}
	if (wn && M.has(e)) return M.get(e);
	if (t) {
		var i = e;
		if (wn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Vn(i)) && (o = lt(i)), M.set(i, o), o;
		}
		var s = !(i.f & 512) && !U && H !== null && !!(H.f & 512), c = (i.f & S) === 0;
		Nn(i) && (s && (i.f |= 512), ut(i)), s && !c && (ft(i), Bn(i));
	}
	if (A?.has(e)) return A.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Bn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (ft(n), Bn(n));
	}
}
function Vn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (M.has(t) || t.f & 2 && Vn(t)) return !0;
	return !1;
}
function Hn(e) {
	var t = U;
	try {
		return U = !0, e();
	} finally {
		U = t;
	}
}
function Un(e) {
	if (!(typeof e != "object" || !e || e instanceof EventTarget)) {
		if (ce in e) Wn(e);
		else if (!Array.isArray(e)) for (let t in e) {
			let n = e[t];
			typeof n == "object" && n && ce in n && Wn(n);
		}
	}
}
function Wn(e, t = /* @__PURE__ */ new Set()) {
	if (typeof e == "object" && e && !(e instanceof EventTarget) && !t.has(e)) {
		t.add(e), e instanceof Date && e.getTime();
		for (let n in e) try {
			Wn(e[n], t);
		} catch {}
		let n = f(e);
		if (n !== Object.prototype && n !== Array.prototype && n !== Map.prototype && n !== Set.prototype && n !== Date.prototype) {
			let t = l(n);
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
var Gn = Symbol("events"), Kn = /* @__PURE__ */ new Set(), qn = /* @__PURE__ */ new Set();
function Jn(e, t, n) {
	(t[Gn] ??= {})[e] = n;
}
function Yn(e) {
	for (var t = 0; t < e.length; t++) Kn.add(e[t]);
	for (var n of qn) n(e);
}
var Xn = null, Zn = !1;
function Qn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Xn = e, Zn || (Zn = !0, setTimeout(() => {
		Zn = !1, Xn = null;
	}));
	var o = 0, c = Xn === e && e[Gn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Gn] = t;
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
		var d = H, f = G;
		W(null), K(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Gn]?.[r];
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
			e[Gn] = t, delete e.currentTarget, W(d), K(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var $n = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function er(e) {
	return $n?.createHTML(e) ?? e;
}
function tr(e) {
	var t = Xt("template");
	return t.innerHTML = er(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function nr(e, t) {
	var n = G;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function rr(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (C) return nr(w, null), w;
		i === void 0 && (i = tr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Gt(i)));
		var t = r || Vt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Gt(t), s = t.lastChild;
			nr(o, s);
		} else nr(t, t);
		return t;
	};
}
function Q(e, t) {
	if (C) {
		var n = G;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = w), xe();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var ir = ["touchstart", "touchmove"];
function ar(e) {
	return ir.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function or(e) {
	let t = 0, n = At(0), r;
	return () => {
		nn() && (Z(n), un(() => (t === 0 && (r = Hn(() => e(() => Lt(n)))), t += 1, () => {
			Je(() => {
				--t, t === 0 && (r?.(), r = void 0, Lt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var sr = te | ne;
function cr(e, t, n, r) {
	new lr(e, t, n, r);
}
var lr = class {
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
	#m = or(() => (this.#p = At(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = G;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = G.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = fn(() => {
			if (C) {
				let e = we(this.#t);
				xe();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, sr), C && (this.#e = w);
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
		Je(r), t && (this.#s = B(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Le(), this.#s !== null && vn(this.#s, () => {
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
					z(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = B(() => e(this.#e)), Je(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = F(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return B(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							z(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(k)) : this.#u === 0 && (this.#e.before(e), this.#c = null, vn(this.#o, () => {
					this.#o = null;
				}), this.#x(k));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = B(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Sn(this.#a, e);
				let t = this.#n.pending;
				this.#o = B(() => t(this.#e));
			} else this.#x(k);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ze(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = G, n = H, r = D;
		K(this.#i), W(this.#i), Ve(this.#i.ctx);
		try {
			return St.ensure(), e();
		} finally {
			K(t), W(n), Ve(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && vn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Je(() => {
			this.#d = !1, this.#p && Pt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Z(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		k?.is_fork ? (this.#a && k.skip_effect(this.#a), this.#o && k.skip_effect(this.#o), this.#s && k.skip_effect(this.#s), k.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (V(this.#a), null), this.#o &&= (V(this.#o), null), this.#s &&= (V(this.#s), null), C && (T(this.#t), Se(), T(Ce()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return B(() => {
						var r = G;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return z(e, this.#i.parent), null;
				}
			}));
		};
		Je(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					z(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => z(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function $(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[me] ??= e.nodeValue) && (e[me] = n, e.nodeValue = `${n}`);
}
function ur(e, t) {
	return fr(e, t);
}
var dr = /* @__PURE__ */ new Map();
function fr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Wt();
	var u = void 0, d = sn(() => {
		var c = r ?? n.appendChild(F());
		cr(c, { pending: () => {} }, (n) => {
			He({});
			var r = D;
			if (s && (r.c = s), a && (i.$$events = a), C && nr(n, null), u = t(n, i) || We(), C && (G.nodes.end = w, w === null || w.nodeType !== 8 || w.data !== "]")) throw ve(), e;
			Ue();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = ar(r);
					for (let e of [n, document]) {
						var a = dr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), dr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Qn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Kn)), qn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = dr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Qn), t.delete(e), t.size === 0 && dr.delete(r)) : t.set(e, i);
			}
			qn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return pr.set(u, d), u;
}
var pr = /* @__PURE__ */ new WeakMap();
function mr(e, t) {
	let n = pr.get(e);
	return n ? (pr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var hr = class {
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
			if (n) bn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (bn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
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
						Sn(r, t), t.append(F()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else V(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), vn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (V(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = k, r = Yt();
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
		} else C && (this.anchor = w), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function gr(e, t, n = !1) {
	var r;
	C && (r = w, xe());
	var i = new hr(e), a = n ? te : 0;
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
	fn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function _r(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		vn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					vr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Jt(d), d.append(u), e.items.clear();
		}
		vr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function vr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ie, Sn(a, document.createDocumentFragment())) : V(t[i], n);
	}
}
var yr;
function br(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = C ? T(/* @__PURE__ */ Gt(u)) : u.appendChild(F());
	}
	C && xe();
	var d = null, f = /* @__PURE__ */ st(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Z(f);
			_.fallback = d, Sr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ie, wr(d, null, c)) : bn(d) : vn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: fn(() => {
			var e = Z(f), r = e.length;
			let o = !1;
			C && we(c) === "[!" != (r === 0) && (c = Ce(), T(c), be(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = k, v = Yt(), y = 0; y < r; y += 1) {
				C && w.nodeType === 8 && w.data === "]" && (c = w, o = !0, be(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Pt(S.v, b), S.i && Pt(S.i, y), v && _.unskip_effect(S.e)) : (S = Cr(l, m ? c : yr ??= F(), b, x, y, a, t, n), m || (S.e.f |= ie), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = B(() => s(c)) : (d = B(() => s(yr ??= F())), d.f |= ie)), r > u.size && ke("", "", ""), C && r > 0 && T(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && be(!0), Z(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, C && (c = w);
}
function xr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Sr(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = xr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (bn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ie, _ === l) wr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Tr(e, d, _), Tr(e, _, y), wr(_, y, n), d = _, p = [], m = [], l = xr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) wr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Tr(e, S.prev, ee.next), Tr(e, d, S), Tr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), wr(_, l, n), Tr(e, _.prev, _.next), Tr(e, _, d === null ? e.effect.first : d.next), Tr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = xr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = xr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (vr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var te = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || te.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && te.push(l), l = xr(l.next);
		var ne = te.length;
		if (ne > 0) {
			var re = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < ne; v += 1) te[v].nodes?.a?.measure();
				for (v = 0; v < ne; v += 1) te[v].nodes?.a?.fix();
			}
			_r(e, te, re);
		}
	}
	a && Je(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Cr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? At(n) : /* @__PURE__ */ jt(n, !1, !1) : null, l = o & 2 ? At(i) : null;
	return {
		v: c,
		i: l,
		e: B(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function wr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Kt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Tr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/elements/actions.js
function Er(e, t, n) {
	cn(() => {
		var r = Hn(() => t(e, n?.()) || {});
		if (n && r?.update) {
			var i = !1, a = {};
			un(() => {
				var e = n();
				Un(e), i && Ee(a, e) && (a = e, r.update(e));
			}), i = !0;
		}
		if (r?.destroy) return () => r.destroy();
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var Dr = [..." 	\n\r\f\xA0\v﻿"];
function Or(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Dr.includes(r[o - 1])) && (s === r.length || Dr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function kr(e, t, n, r, i, a) {
	var o = e[fe];
	if (C || o !== n || o === void 0) {
		var s = Or(n, r, a);
		C && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[fe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Ar = Symbol("is custom element"), jr = Symbol("is html"), Mr = ge ? "link" : "LINK";
function Nr(e, t, n, r) {
	var i = Pr(e);
	if (!(C && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Mr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ue] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Ir(e).has(t) ? e[t] = n : (!C || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Pr(e) {
	return e[de] ??= {
		[Ar]: e.nodeName.includes("-"),
		[jr]: e.namespaceURI === n
	};
}
var Fr = /* @__PURE__ */ new Map();
function Ir(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Fr.get(t);
	if (n) return n;
	Fr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// kit/plugin.svelte.js
function Lr(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ N(zt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return Z(this.#e);
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
			if (!document.querySelector(`link[href="${n}"]`)) {
				let e = document.createElement("link");
				e.rel = "stylesheet", e.href = n, document.head.append(e);
			}
			this.style.display = "contents", this.app ||= ur(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && mr(this.app), this.app = null;
		}
	});
}
function Rr(e, t) {
	let n = zt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			It(r);
		}
	}), r = /* @__PURE__ */ N(0);
	an(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, It(r);
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
	return an(() => {
		let n = e(), c = t();
		Z(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function zr(e) {
	let t = zt({});
	return an(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
function Br(e, t) {
	let n = t.api.slot(e, t.slot, t.context, t.target);
	return {
		update(r) {
			n && r.slot === t.slot && JSON.stringify(r.target) === JSON.stringify(t.target) ? n.context(r.context) : (n?.destroy(), n = r.api.slot(e, r.slot, r.context, r.target)), t = r;
		},
		destroy() {
			n?.destroy();
		}
	};
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/format.js
function Vr(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function Hr(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
function Ur(e) {
	return !e || !e.meta ? {} : typeof e.meta == "string" ? JSON.parse(e.meta) : e.meta;
}
function Wr(e) {
	switch (e) {
		case "done":
		case "running":
		case "closed": return "ok";
		case "needs_human":
		case "in_progress": return "warn";
		case "failed":
		case "cancelled":
		case "crashed":
		case "stuck": return "bad";
		default: return "";
	}
}
function Gr(e) {
	let t = Ur(e), n = [], r = t.github || {}, i = r.repo || "elee1766/jobshop";
	r.issue && n.push({
		label: "issue #" + r.issue,
		href: `https://github.com/${i}/issues/${r.issue}`
	}), r.pr && n.push({
		label: "PR #" + r.pr,
		href: `https://github.com/${i}/pull/${r.pr}`
	});
	let a = t.linear || {}, o = a.issue || a.identifier, s = a.url || (a.workspace && o ? `https://linear.app/${encodeURIComponent(a.workspace)}/issue/${encodeURIComponent(o)}` : "");
	return s && Kr(s) && n.push({
		label: o || "Linear",
		href: s
	}), n;
}
function Kr(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// plugins/work/Work.svelte
var qr = /* @__PURE__ */ rr("<button> </button>"), Jr = /* @__PURE__ */ rr("<div class=\"card err\"> </div>"), Yr = /* @__PURE__ */ rr("<tr class=\"click\"><td class=\"mono c-id\"> </td><td class=\"c-title\"> </td><td class=\"muted c-time\"> </td></tr>"), Xr = /* @__PURE__ */ rr("<h2 class=\"st\">Batches</h2> <div class=\"card scroll\"><table class=\"stack\"><tbody></tbody></table></div>", 1), Zr = /* @__PURE__ */ rr("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"svelte-1k64nxg\"> </a>"), Qr = /* @__PURE__ */ rr("<tr class=\"click\"><td class=\"c-title\"> <div class=\"faint mono\"> </div></td><td class=\"c-id\"><span> </span></td><td class=\"muted\"> </td><td class=\"links svelte-1k64nxg\"></td><td class=\"muted c-time\"> </td></tr>"), $r = /* @__PURE__ */ rr("<tr><td colspan=\"5\" class=\"faint\">No runs.</td></tr>"), ei = /* @__PURE__ */ rr("<!> <h2 class=\"st\">Runs</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"runs\"><thead><tr><th>Run</th><th>State</th><th>Workflow</th><th>Links</th><th>Updated</th></tr></thead><tbody></tbody></table></div>", 1), ti = /* @__PURE__ */ rr("<header class=\"vh head svelte-1k64nxg\"><h1>Work</h1> <div class=\"chips actions svelte-1k64nxg\" role=\"group\" aria-label=\"state filter\"></div> <div class=\"toolbar svelte-1k64nxg\" data-testid=\"work-toolbar\"></div></header> <div class=\"body svelte-1k64nxg\"><div class=\"pad main svelte-1k64nxg\"><!></div> <aside class=\"pad aside svelte-1k64nxg\" data-testid=\"work-aside\"></aside></div>", 1);
function ni(e, t) {
	He(t, !0);
	let n = zr(() => t.api), r = (e) => t.api.go(e), i = [
		"",
		"running",
		"needs_human",
		"done",
		"cancelled"
	], a = Rr(() => t.api, () => "api/work" + (n.state ? "?state=" + encodeURIComponent(n.state) : "")), o = /* @__PURE__ */ ot(() => ({ state: n.state || "" })), s = /* @__PURE__ */ ot(() => ({
		state: n.state || "",
		work: ze(a.data)
	}));
	var c = ti(), l = qt(c), u = R(I(l), 2);
	br(u, 20, () => i, (e) => e, (e, t) => {
		var i = qr();
		let a;
		var o = L(i, !0);
		dn((e) => {
			a = kr(i, 1, "", null, a, { on: n.state === t }), $(o, e);
		}, [() => t ? t.replace("_", " ") : "all"]), Jn("click", i, () => r({ state: t })), Q(e, i);
	}), E(u), Er(R(u, 2), (e, t) => Br?.(e, t), () => ({
		api: t.api,
		slot: "work:toolbar",
		context: Z(o)
	})), E(l);
	var d = R(l, 2), f = I(d), p = I(f), m = (e) => {
		var t = Jr(), n = L(t, !0);
		dn(() => $(n, a.error)), Q(e, t);
	}, h = (e) => {
		var t = ei(), n = qt(t), i = (e) => {
			var t = Xr(), n = R(qt(t), 2), i = I(n), o = I(i);
			br(o, 21, () => a.data.batches, (e) => e.id, (e, t) => {
				var n = Yr(), i = I(n), a = L(i, !0), o = R(i), s = L(o, !0), c = L(R(o), !0);
				E(n), dn((e) => {
					$(a, Z(t).id), $(s, Z(t).title), $(c, e);
				}, [() => Vr(Z(t).updated_at)]), Jn("click", n, () => r({ ticket: Z(t).id })), Q(e, n);
			}), E(o), E(i), E(n), Q(e, t);
		};
		gr(n, (e) => {
			a.data.batches.length && e(i);
		});
		var o = R(n, 4), s = I(o), c = R(I(s));
		br(c, 21, () => a.data.runs, (e) => e.id, (e, t) => {
			let n = /* @__PURE__ */ ot(() => Hr(Z(t)));
			var i = Qr(), a = I(i), o = I(a, !0), s = L(R(o));
			E(a);
			var c = R(a), l = I(c), u = L(l, !0);
			E(c);
			var d = R(c), f = L(d, !0), p = R(d);
			br(p, 21, () => Gr(Z(t)), (e) => e.href, (e, t) => {
				var n = Zr(), r = L(n, !0);
				dn(() => {
					Nr(n, "href", Z(t).href), $(r, Z(t).label);
				}), Jn("click", n, (e) => e.stopPropagation()), Q(e, n);
			}), E(p);
			var m = L(R(p), !0);
			E(i), dn((e, r) => {
				$(o, Z(t).title), $(s, `${Z(t).id ?? ""} · ${Z(t).project ?? ""}`), kr(l, 1, `pill ${e ?? ""}`, "svelte-1k64nxg"), $(u, Z(n).state || Z(t).status), $(f, Z(n).workflow || ""), $(m, r);
			}, [() => Wr(Z(n).state), () => Vr(Z(t).updated_at)]), Jn("click", i, () => r({ ticket: Z(t).id })), Q(e, i);
		}, (e) => {
			Q(e, $r());
		}), E(c), E(s), E(o), Q(e, t);
	};
	gr(p, (e) => {
		a.error ? e(m) : a.data && e(h, 1);
	}), E(f), Er(R(f, 2), (e, t) => Br?.(e, t), () => ({
		api: t.api,
		slot: "work:aside",
		context: Z(s)
	})), E(d), Q(e, c), Ue();
}
// plugins/work/index.js
Yn(["click"]), Lr("desk-work-view", ni);
