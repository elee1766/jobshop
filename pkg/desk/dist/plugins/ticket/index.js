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
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, C = 1 << 25, ee = 65536, te = 1 << 19, ne = 1 << 20, w = 1 << 25, re = 1 << 21, ie = 1 << 22, ae = 1 << 23, oe = Symbol("$state"), se = Symbol("component"), ce = Symbol("legacy props"), le = Symbol(""), ue = Symbol("attributes"), de = Symbol("class"), fe = Symbol("style"), pe = Symbol("text"), me = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), he = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function ge() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function T(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function _e() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var E = !1;
function ve(e) {
	E = e;
}
var D;
function ye(t) {
	if (t === null) throw T(), e;
	return D = t;
}
function be() {
	return ye(/* @__PURE__ */ F(D));
}
function O(t) {
	if (E) {
		if (/* @__PURE__ */ F(D) !== null) throw T(), e;
		D = t;
	}
}
function xe(e = 1) {
	if (E) {
		for (var t = e, n = D; t--;) n = /* @__PURE__ */ F(n);
		D = n;
	}
}
function Se(e = !0) {
	for (var t = 0, n = D;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ F(n);
		e && n.remove(), n = i;
	}
}
function k(t) {
	if (!t || t.nodeType !== 8) throw T(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Ce(e) {
	return e === this.v;
}
function A(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function we(e) {
	return !A(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function Te() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Ee(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function De(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Oe() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function ke(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Ae() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function je(e) {
	throw Error("https://svelte.dev/e/props_invalid_value");
}
function Me() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ne() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Pe() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Fe() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var Ie = null;
function Le(e) {
	Ie = e;
}
function Re(e, t = !1, n) {
	Ie = {
		p: Ie,
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
	var t = Ie, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) on(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, Ie = t.p, Be(e);
}
function Be(e = {}) {
	return s(e, se, { value: !0 }), e;
}
function Ve() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var He = [];
function Ue() {
	var e = He;
	He = [], h(e);
}
function We(e) {
	if (He.length === 0 && !pt) {
		var t = He;
		queueMicrotask(() => {
			t === He && Ue();
		});
	}
	He.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ge = ~(v | y | _);
function j(e, t) {
	e.f = e.f & Ge | t;
}
function Ke(e) {
	e.f & 512 || e.deps === null ? j(e, _) : j(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function qe(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), j(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Je(e) {
	var t = B, n = V;
	Dn(null), On(null);
	try {
		return e();
	} finally {
		Dn(t), On(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ye(e, t, n, r) {
	let i = Ve() ? $e : rt;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = V, c = Xe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				Qt(e, s);
			}
			Ze();
		}
	}
	var d = Qe();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ tt(e))).then(u).catch((e) => Qt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Ze());
	}) : f();
}
function Xe() {
	var e = V, t = B, n = Ie, r = M;
	return function(i = !0) {
		On(e), Dn(t), Le(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Ze(e = !0) {
	On(null), Dn(null), Le(null), e && M?.deactivate();
}
function Qe() {
	var e = V, t = e.b, n = M, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function $e(e) {
	var n = 2 | v;
	return V !== null && (V.f |= te), {
		ctx: Ie,
		deps: null,
		effects: null,
		equals: Ce,
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
var et = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function tt(e, n, r) {
	let i = V;
	i === null && Te();
	var a = void 0, o = Ot(t), s = !B, c = /* @__PURE__ */ new Set();
	return cn(() => {
		var t = V, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== me && n.reject(e);
			}).finally(Ze);
		} catch (e) {
			n.reject(e), Ze();
		}
		var r = M;
		if (s) {
			if (t.f & 32768) var l = Qe();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(et);
			else for (let e of c.values()) e.reject(et);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== et && (r.activate(), t ? (o.f |= ae, Pt(o, t)) : (o.f & 8388608 && (o.f ^= ae), Pt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), rn(() => {
		for (let e of c) e.reject(et);
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
function nt(e) {
	let t = /* @__PURE__ */ $e(e);
	return An(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function rt(e) {
	let t = /* @__PURE__ */ $e(e);
	return t.equals = we, t;
}
function it(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) hn(t[n]);
	}
}
function at(e) {
	var n, r = V, i = e.parent;
	if (!wn && i !== null && e.v !== t && i.f & 24576) return ge(), e.v;
	On(i);
	try {
		it(e), n = Hn(e);
	} finally {
		On(r);
	}
	return n;
}
function ot(e) {
	var t = at(e);
	!e.equals(t) && (e.wv = zn(), (!M?.is_fork || e.deps === null) && (M === null ? e.v = t : (M.capture(e, t, !0), ut?.capture(e, t, !0)), e.deps === null)) ? j(e, _) : wn || (dt === null ? Ke(e) : (nn() || M?.is_fork) && dt.set(e, t));
}
function st(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Je(() => {
		t.ac.abort(me), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Gn(t, 0), pn(t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Kn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var lt = null, M = null, ut = null, dt = null, ft = null, pt = !1, N = !1, mt = null, ht = null, gt = 0, _t = 1, vt = class e {
	id = _t++;
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
		lt === null ? lt = this : (lt.next = this, this.prev = lt), lt = this;
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
			for (var r of n.d) j(r, v), t(r);
			for (r of n.m) j(r, y), t(r);
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
		for (let [e, t] of this.#s) j(e, t), this.schedule(e);
		this.apply();
		for (var t = mt = [], n = [], r = ht = []; this.#a.length > 0;) {
			gt++ > 1e3 && (this.#b(), yt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw wt(e), this.#d() || this.discard(), t;
			}
		}
		if (M = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (mt = null, ht = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Ct(e, t);
			r.length > 0 && M.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), ut = this, xt(n), xt(t), ut = null, this.#i?.resolve();
			var o = M;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Et.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : Bn(r) && (i & 16 && this.add_dirty_reaction(r, y), Kn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), j(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), M = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) qe(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), dt?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		M = this;
	}
	deactivate() {
		M = null, dt = null;
	}
	flush() {
		try {
			N = !0, M = this, this.#m();
		} finally {
			gt = 0, ft = null, mt = null, ht = null, N = !1, M = null, dt = null, Et.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(et);
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
		return (this.#i ??= g()).promise;
	}
	static ensure() {
		if (M === null) {
			let t = M = new e();
			!N && We(() => {
				t.started || t.flush();
			});
		}
		return M;
	}
	apply() {
		dt = null;
	}
	schedule(e) {
		ft = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? lt = e : t.prev = e, this.linked = !1;
		}
	}
};
function yt() {
	try {
		Ae();
	} catch (e) {
		Qt(e, ft);
	}
}
var bt = null;
function xt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Bn(r) && (bt = /* @__PURE__ */ new Set(), Kn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && _n(r), bt?.size > 0)) {
				Et.clear();
				for (let e of bt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) bt.has(n) && (bt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Kn(n);
					}
				}
				bt.clear();
			}
		}
		bt = null;
	}
}
function St(e) {
	M.schedule(e);
}
function Ct(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), j(e, _);
		for (var n = e.first; n !== null;) Ct(n, t), n = n.next;
	}
}
function wt(e) {
	j(e, _);
	for (var t = e.first; t !== null;) wt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Tt = /* @__PURE__ */ new Set(), Et = /* @__PURE__ */ new Map(), Dt = !1;
function Ot(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Ce,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function kt(e, t) {
	let n = Ot(e, t);
	return An(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function At(e, t = !1, n = !0) {
	let r = Ot(e);
	return t || (r.equals = we), r;
}
function jt(e, t, n = !1) {
	return B !== null && (!En || B.f & 131072) && Ve() && B.f & 4325394 && (kn === null || !kn.has(e)) && Pe(), Pt(e, n ? zt(t) : t, ht);
}
var Mt = null, Nt = 0;
function Pt(e, t, n = null) {
	if (!e.equals(t)) {
		wn ? Et.set(e, t) : Et.has(e) || Et.set(e, e.v);
		var r = vt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && at(t), dt === null && Ke(t);
		}
		e.wv = zn(), Mt = null, Nt = 0, Rt(e, v, n), Mt = null, Ve() && V !== null && V.f & 1024 && !(V.f & 96) && (Nn === null ? Pn([e]) : Nn.push(e)), !r.is_fork && Tt.size > 0 && !Dt && Ft();
	}
	return t;
}
function Ft() {
	Dt = !1;
	for (let e of Tt) {
		e.f & 1024 && j(e, y);
		let t;
		try {
			t = Bn(e);
		} catch {
			t = !0;
		}
		t && Kn(e);
	}
	Tt.clear();
}
function It(e, t = 1) {
	var n = H(e), r = t === 1 ? n++ : n--;
	return jt(e, n), r;
}
function Lt(e) {
	jt(e, e.v + 1);
}
function Rt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ve(), a = r.length;
		if (Nt += a, Nt > 1e5 && Mt === null && (Mt = /* @__PURE__ */ new Set()), Mt !== null) {
			if (Mt.has(e)) return;
			Mt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== V) {
				var l = (c & v) === 0;
				if (l && j(s, t), c & 131072) Tt.add(s);
				else if (c & 2) {
					var u = s;
					dt?.delete(u), Rt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && bt !== null && bt.add(d), n === null ? St(d) : n.push(d);
				}
			}
		}
	}
}
function zt(e) {
	if (typeof e != "object" || !e || oe in e || se in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ kt(0), s = null, l = Ln, p = (e) => {
		if (Ln === l) return e();
		var t = B, n = Ln;
		Dn(null), Rn(l);
		var r = e();
		return Dn(t), Rn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ kt(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Me();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ kt(n.value, s);
				return i.set(t, e), e;
			}) : jt(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ kt(t, s));
					i.set(n, e), Lt(o);
				}
			} else jt(r, t), Lt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === oe) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ kt(zt(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = H(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = H(a);
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
			if (n === oe) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (V !== null || B !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ kt(a ? zt(e[n]) : t, s)), i.set(n, r)), H(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ kt(t, s)), i.set(f + "", m)) : jt(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ kt(void 0, s)), jt(u, zt(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => zt(r));
				jt(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && jt(_, v + 1);
				}
				Lt(o);
			}
			return !0;
		},
		ownKeys(e) {
			H(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Ne();
		}
	});
}
var Bt, Vt, Ht, Ut;
function Wt() {
	if (Bt === void 0) {
		Bt = window, Vt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Ht = c(t, "firstChild").get, Ut = c(t, "nextSibling").get, p(e) && (e[de] = void 0, e[ue] = null, e[fe] = void 0, e.__e = void 0), p(n) && (n[pe] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Gt(e) {
	return Ht.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function F(e) {
	return Ut.call(e);
}
function I(e, t) {
	if (!E) return /* @__PURE__ */ Gt(e);
	var n = /* @__PURE__ */ Gt(D);
	if (n === null) n = D.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), ye(r), r;
	}
	return t && Xt(n), ye(n), n;
}
function Kt(e, t = !1) {
	if (!E) {
		var n = /* @__PURE__ */ Gt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ F(n) : n;
	}
	if (t) {
		if (D?.nodeType !== 3) {
			var r = P();
			return D?.before(r), ye(r), r;
		}
		Xt(D);
	}
	return D;
}
function L(e, t = !1) {
	if (!E) return /* @__PURE__ */ Gt(e);
	var n = I(e, t);
	return O(e), n;
}
function R(e, t = 1, n = !1) {
	let r = E ? D : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ F(r);
	if (!E) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), ye(a), a;
		}
		Xt(r);
	}
	return ye(r), r;
}
function qt(e) {
	e.textContent = "";
}
function Jt() {
	return !1;
}
function Yt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Xt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Zt(e) {
	var t = V;
	if (t === null) return B.f |= ae, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	Qt(e, t);
}
function Qt(t, n) {
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
	V === null && (B === null && ke(e), Oe()), wn && De(e);
}
function en(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function tn(e, t) {
	var n = V;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: Ie,
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
	M?.register_created_effect(r);
	var i = r;
	if (e & 4) mt === null ? vt.ensure().schedule(r) : mt.push(r);
	else if (t !== null) {
		try {
			Kn(r);
		} catch (e) {
			throw hn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ee));
	}
	if (i !== null && (i.parent = n, n !== null && en(i, n), B !== null && B.f & 2 && !(e & 64))) {
		var a = B;
		(a.effects ??= []).push(i);
	}
	return r;
}
function nn() {
	return B !== null && !En;
}
function rn(e) {
	let t = tn(8, null);
	return j(t, _), t.teardown = e, t;
}
function an(e) {
	$t("$effect");
	var t = V.f;
	if (!B && t & 32 && Ie !== null && !Ie.i) {
		var n = Ie;
		(n.e ??= []).push(e);
	} else return on(e);
}
function on(e) {
	return tn(4 | ne, e);
}
function sn(e) {
	vt.ensure();
	let t = tn(64 | te, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? vn(t, () => {
			hn(t), n(void 0);
		}) : (hn(t), n(void 0));
	});
}
function cn(e) {
	return tn(ie | te, e);
}
function ln(e, t = 0) {
	return tn(8 | t, e);
}
function z(e, t = [], n = [], r = []) {
	Ye(r, t, n, (t) => {
		tn(8, () => {
			e(...t.map(H));
		});
	});
}
function un(e, t = 0) {
	return tn(16 | t, e);
}
function dn(e) {
	return tn(32 | te, e);
}
function fn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = wn, r = B;
		Tn(!0), Dn(null);
		try {
			t.call(null);
		} catch (t) {
			Qt(t, e.parent);
		} finally {
			Tn(n), Dn(r);
		}
	}
}
function pn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Je(() => {
			e.abort(me);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : hn(n, t), n = r;
	}
}
function mn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || hn(t), t = n;
	}
}
function hn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (gn(e.nodes.start, e.nodes.end), n = !0), e.f |= C, pn(e, t && !n), Gn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	fn(e), e.f ^= C, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && _n(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function gn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ F(e);
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
		n && hn(e), t && t();
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
		e.f ^= b, e.f & 1024 || (j(e, v), vt.ensure().schedule(e));
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
		var i = n === r ? null : /* @__PURE__ */ F(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Cn = null, wn = !1;
function Tn(e) {
	wn = e;
}
var B = null, En = !1;
function Dn(e) {
	B = e;
}
var V = null;
function On(e) {
	V = e;
}
var kn = null;
function An(e) {
	B !== null && (B.f & 2097152 || B.f & 2) && (kn ??= /* @__PURE__ */ new Set()).add(e);
}
var jn = null, Mn = 0, Nn = null;
function Pn(e) {
	Nn = e;
}
var Fn = 1, In = 0, Ln = In;
function Rn(e) {
	Ln = e;
}
function zn() {
	return ++Fn;
}
function Bn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Bn(a) && ot(a), a.wv > e.wv) return !0;
		}
		t & 512 && dt === null && j(e, _);
	}
	return !1;
}
function Vn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(kn !== null && kn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Vn(a, t, !1) : t === a && (n ? j(a, v) : a.f & 1024 && j(a, y), St(a));
	}
}
function Hn(e) {
	var t = jn, n = Mn, r = Nn, i = B, a = kn, o = Ie, s = En, c = Ln, l = e.f;
	jn = null, Mn = 0, Nn = null, B = l & 96 ? null : e, kn = null, Le(e.ctx), En = !1, Ln = ++In, e.ac !== null && (Je(() => {
		e.ac.abort(me);
	}), e.ac = null);
	try {
		e.f |= re;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Un(e);
		if (Ve() && Nn !== null && !En && f !== null && !(e.f & 6146)) for (var p = 0; p < Nn.length; p++) Vn(Nn[p], e);
		if (i !== null && i !== e) {
			if (In++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = In;
			if (t !== null) for (let e of t) e.rv = In;
			Nn !== null && (r === null ? r = Nn : r.push(...Nn));
		}
		return e.f & 8388608 && (e.f ^= ae), d;
	} catch (t) {
		return Un(e), Zt(t);
	} finally {
		e.f ^= re, jn = t, Mn = n, Nn = r, B = i, kn = a, Le(o), En = s, Ln = c;
	}
}
function Un(e) {
	var t = e.deps, n = M?.is_fork;
	if (jn !== null) {
		var r;
		if (n || Gn(e, Mn), t !== null && Mn > 0) for (t.length = Mn + jn.length, r = 0; r < jn.length; r++) t[Mn + r] = jn[r];
		else e.deps = t = jn;
		if (nn() && e.f & 512) for (r = Mn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Mn < t.length && (Gn(e, Mn), t.length = Mn);
	return t;
}
function Wn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (jn === null || !a.call(jn, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ke(c), c.ac !== null && Je(() => {
			c.ac.abort(me), c.ac = null, j(c, v);
		}), st(c), Gn(c, 0);
	}
}
function Gn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Wn(e, n[r]);
}
function Kn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		j(e, _);
		var n = V;
		V = e;
		try {
			t & 16777232 ? mn(e) : pn(e), fn(e);
			var r = Hn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Fn;
		} finally {
			V = n;
		}
	}
}
function H(e) {
	var t = !!(e.f & 2);
	if (Cn?.add(e), B !== null && !En && !(V !== null && V.f & 16384) && (kn === null || !kn.has(e))) {
		var n = B.deps;
		if (B.f & 2097152) e.rv < In && (e.rv = In, jn === null && n !== null && n[Mn] === e ? Mn++ : jn === null ? jn = [e] : jn.push(e));
		else {
			B.deps ??= [], a.call(B.deps, e) || B.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [B] : a.call(r, B) || r.push(B);
		}
	}
	if (wn && Et.has(e)) return Et.get(e);
	if (t) {
		var i = e;
		if (wn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Jn(i)) && (o = at(i)), Et.set(i, o), o;
		}
		var s = !(i.f & 512) && !En && B !== null && !!(B.f & 512), c = (i.f & S) === 0;
		Bn(i) && (s && (i.f |= 512), ot(i)), s && !c && (ct(i), qn(i));
	}
	if (dt?.has(e)) return dt.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function qn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (ct(n), qn(n));
	}
}
function Jn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Et.has(t) || t.f & 2 && Jn(t)) return !0;
	return !1;
}
function Yn(e) {
	var t = En;
	try {
		return En = !0, e();
	} finally {
		En = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Xn = Symbol("events"), Zn = /* @__PURE__ */ new Set(), Qn = /* @__PURE__ */ new Set();
function $n(e, t, n) {
	(t[Xn] ??= {})[e] = n;
}
function er(e) {
	for (var t = 0; t < e.length; t++) Zn.add(e[t]);
	for (var n of Qn) n(e);
}
var tr = null, nr = !1;
function rr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	tr = e, nr || (nr = !0, setTimeout(() => {
		nr = !1, tr = null;
	}));
	var o = 0, c = tr === e && e[Xn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Xn] = t;
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
		var d = B, f = V;
		Dn(null), On(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Xn]?.[r];
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
			e[Xn] = t, delete e.currentTarget, Dn(d), On(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var ir = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function ar(e) {
	return ir?.createHTML(e) ?? e;
}
function or(e) {
	var t = Yt("template");
	return t.innerHTML = ar(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function sr(e, t) {
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
		if (E) return sr(D, null), D;
		i === void 0 && (i = or(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Gt(i)));
		var t = r || Vt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Gt(t), s = t.lastChild;
			sr(o, s);
		} else sr(t, t);
		return t;
	};
}
function cr() {
	if (E) return sr(D, null), D;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = P();
	return e.append(t, n), sr(t, n), e;
}
function W(e, t) {
	if (E) {
		var n = V;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = D), be();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var lr = ["touchstart", "touchmove"];
function ur(e) {
	return lr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function dr(e) {
	let t = 0, n = Ot(0), r;
	return () => {
		nn() && (H(n), ln(() => (t === 0 && (r = Yn(() => e(() => Lt(n)))), t += 1, () => {
			We(() => {
				--t, t === 0 && (r?.(), r = void 0, Lt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var fr = ee | te;
function pr(e, t, n, r) {
	new mr(e, t, n, r);
}
var mr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = E ? D : null;
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
	#m = dr(() => (this.#p = Ot(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = V;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = V.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = un(() => {
			if (E) {
				let e = k(this.#t);
				be();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, fr), E && (this.#e = D);
	}
	#h() {
		try {
			this.#a = dn(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		We(r), t && (this.#s = dn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? _e() : (t = !0, n && Fe(), this.#s !== null && vn(this.#s, () => {
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
					Qt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | C)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = dn(() => e(this.#e)), We(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = P(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return dn(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							Qt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(M)) : this.#u === 0 && (this.#e.before(e), this.#c = null, vn(this.#o, () => {
					this.#o = null;
				}), this.#x(M));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = dn(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Sn(this.#a, e);
				let t = this.#n.pending;
				this.#o = dn(() => t(this.#e));
			} else this.#x(M);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		qe(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = V, n = B, r = Ie;
		On(this.#i), Dn(this.#i), Le(this.#i.ctx);
		try {
			return vt.ensure(), e();
		} finally {
			On(t), Dn(n), Le(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && vn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, We(() => {
			this.#d = !1, this.#p && Pt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), H(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		M?.is_fork ? (this.#a && M.skip_effect(this.#a), this.#o && M.skip_effect(this.#o), this.#s && M.skip_effect(this.#s), M.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (hn(this.#a), null), this.#o &&= (hn(this.#o), null), this.#s &&= (hn(this.#s), null), E && (ye(this.#t), xe(), ye(Se()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return dn(() => {
						var r = V;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return Qt(e, this.#i.parent), null;
				}
			}));
		};
		We(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					Qt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => Qt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function G(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[pe] ??= e.nodeValue) && (e[pe] = n, e.nodeValue = `${n}`);
}
function hr(e, t) {
	return _r(e, t);
}
var gr = /* @__PURE__ */ new Map();
function _r(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Wt();
	var u = void 0, d = sn(() => {
		var c = r ?? n.appendChild(P());
		pr(c, { pending: () => {} }, (n) => {
			Re({});
			var r = Ie;
			if (s && (r.c = s), a && (i.$$events = a), E && sr(n, null), u = t(n, i) || Be(), E && (V.nodes.end = D, D === null || D.nodeType !== 8 || D.data !== "]")) throw T(), e;
			ze();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = ur(r);
					for (let e of [n, document]) {
						var a = gr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), gr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, rr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Zn)), Qn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = gr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, rr), t.delete(e), t.size === 0 && gr.delete(r)) : t.set(e, i);
			}
			Qn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return vr.set(u, d), u;
}
var vr = /* @__PURE__ */ new WeakMap();
function yr(e, t) {
	let n = vr.get(e);
	return n ? (vr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var br = class {
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
				r && (hn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Sn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else hn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), vn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (hn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = M, r = Jt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = P();
				i.append(a), this.#n.set(e, {
					effect: dn(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, dn(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else E && (this.anchor = D), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function K(e, t, n = !1) {
	var r;
	E && (r = D, be());
	var i = new br(e), a = n ? ee : 0;
	function o(e, t) {
		if (E) {
			var n = k(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Se();
				ye(a), i.anchor = a, ve(!1), i.ensure(e, t), ve(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	un(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function xr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		vn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Sr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			qt(d), d.append(u), e.items.clear();
		}
		Sr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Sr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= w, Sn(a, document.createDocumentFragment())) : hn(t[i], n);
	}
}
var Cr;
function wr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = E ? ye(/* @__PURE__ */ Gt(u)) : u.appendChild(P());
	}
	E && be();
	var d = null, f = /* @__PURE__ */ rt(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = H(f);
			_.fallback = d, Er(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= w, Or(d, null, c)) : bn(d) : vn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: un(() => {
			var e = H(f), r = e.length;
			let o = !1;
			E && k(c) === "[!" != (r === 0) && (c = Se(), ye(c), ve(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = M, v = Jt(), y = 0; y < r; y += 1) {
				E && D.nodeType === 8 && D.data === "]" && (c = D, o = !0, ve(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Pt(S.v, b), S.i && Pt(S.i, y), v && _.unskip_effect(S.e)) : (S = Dr(l, m ? c : Cr ??= P(), b, x, y, a, t, n), m || (S.e.f |= w), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = dn(() => s(c)) : (d = dn(() => s(Cr ??= P())), d.f |= w)), r > u.size && Ee("", "", ""), E && r > 0 && ye(Se()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && ve(!0), H(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, E && (c = D);
}
function Tr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Er(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = Tr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (bn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= w, _ === l) Or(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), kr(e, d, _), kr(e, _, y), Or(_, y, n), d = _, p = [], m = [], l = Tr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Or(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					kr(e, S.prev, C.next), kr(e, d, S), kr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Or(_, l, n), kr(e, _.prev, _.next), kr(e, _, d === null ? e.effect.first : d.next), kr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Tr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Tr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Sr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var ee = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || ee.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && ee.push(l), l = Tr(l.next);
		var te = ee.length;
		if (te > 0) {
			var ne = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) ee[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) ee[v].nodes?.a?.fix();
			}
			xr(e, ee, ne);
		}
	}
	a && We(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Dr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Ot(n) : /* @__PURE__ */ At(n, !1, !1) : null, l = o & 2 ? Ot(i) : null;
	return {
		v: c,
		i: l,
		e: dn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Or(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ F(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function kr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var Ar = [..." 	\n\r\f\xA0\v﻿"];
function jr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Ar.includes(r[o - 1])) && (s === r.length || Ar.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Mr(e, t, n, r, i, a) {
	var o = e[de];
	if (E || o !== n || o === void 0) {
		var s = jr(n, r, a);
		E && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[de] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Nr = Symbol("is custom element"), Pr = Symbol("is html"), Fr = he ? "link" : "LINK";
function Ir(e, t, n, r) {
	var i = Lr(e);
	if (!(E && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Fr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[le] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && zr(e).has(t) ? e[t] = n : (!E || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Lr(e) {
	return e[ue] ??= {
		[Nr]: e.nodeName.includes("-"),
		[Pr]: e.namespaceURI === n
	};
}
var Rr = /* @__PURE__ */ new Map();
function zr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Rr.get(t);
	if (n) return n;
	Rr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Br = !1;
function Vr(e) {
	var t = Br;
	try {
		return Br = !1, [e(), Br];
	} finally {
		Br = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function Hr(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, l = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ $e(r), H(u)) : (l && (l = !1, s = o ? Yn(r) : r), s);
	let f;
	if (a) {
		var p = oe in e || ce in e;
		f = c(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Vr(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && je(t), f(m)));
	var g = i ? () => {
		var n = e[t];
		return n === void 0 ? d() : (l = !0, n);
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
	var v = !1, y = (n & 1 ? $e : rt)(() => (v = !1, g()));
	a && H(y);
	var b = V;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? H(y) : i && a ? zt(e) : e;
			return jt(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return wn && v || b.f & 16384 ? y.v : H(y);
	});
}
// kit/plugin.svelte.js
function Ur(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ kt(zt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return H(this.#e);
		}
		set props(e) {
			jt(this.#e, e, !0);
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
			this.style.display = "contents", this.app ||= hr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && yr(this.app), this.app = null;
		}
	});
}
function Wr(e, t) {
	let n = zt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			It(r);
		}
	}), r = /* @__PURE__ */ kt(0);
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
		H(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/marked/lib/marked.esm.js
function Gr() {
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
var Kr = Gr();
function qr(e) {
	Kr = e;
}
var Jr = { exec: () => null };
function Yr(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function q(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(J.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var Xr = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), J = {
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
	nextBulletRegex: Yr((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Yr((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Yr((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Yr((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Yr((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Yr((e) => RegExp(`^ {0,${e}}(?:</?(?:${di})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Yr((e) => RegExp(`^ {0,${e}}>`))
}, Zr = /^(?:[ \t]*(?:\n|$))+/, Qr = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, $r = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, ei = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, ti = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, ni = / {0,3}(?:[*+-]|\d{1,9}[.)])/, ri = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, ii = q(ri).replace(/bull/g, ni).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), ai = q(ri).replace(/bull/g, ni).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), oi = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, si = /^[^\n]+/, ci = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, li = q(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", ci).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), ui = q(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, ni).getRegex(), di = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", fi = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, pi = q("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", fi).replace("tag", di).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), mi = (e) => q(oi).replace("hr", ei).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", di).getRegex(), hi = mi(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), gi = mi(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), _i = {
	blockquote: q(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", gi).getRegex(),
	code: Qr,
	def: li,
	fences: $r,
	heading: ti,
	hr: ei,
	html: pi,
	lheading: ii,
	list: ui,
	newline: Zr,
	paragraph: hi,
	table: Jr,
	text: si
}, vi = q("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", ei).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", di).getRegex(), yi = {
	..._i,
	lheading: ai,
	table: vi,
	paragraph: q(oi).replace("hr", ei).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", vi).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", di).getRegex()
}, bi = {
	..._i,
	html: q("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", fi).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: Jr,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: q(oi).replace("hr", ei).replace("heading", " *#{1,6} *[^\n]").replace("lheading", ii).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, xi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Si = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Ci = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, wi = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, Ti = /[\p{P}\p{S}]/u, Ei = /[\s\p{P}\p{S}]/u, Di = /[^\s\p{P}\p{S}]/u, Oi = q(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Ei).getRegex(), ki = /[\p{Pi}\p{Ps}"']/u, Ai = /(?!~)[\p{P}\p{S}]/u, ji = /(?!~)[\s\p{P}\p{S}]/u, Mi = /(?:[^\s\p{P}\p{S}]|~)/u, Ni = q(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", Xr ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), Pi = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, Fi = q(Pi, "u").replace(/punct/g, Ti).getRegex(), Ii = q(Pi, "u").replace(/punct/g, Ai).getRegex(), Li = q(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, ki).replace(/punct/g, Ti).getRegex(), Ri = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", zi = q(Ri, "gu").replace(/notPunctSpace/g, Di).replace(/punctSpace/g, Ei).replace(/punct/g, Ti).getRegex(), Bi = q(Ri, "gu").replace(/notPunctSpace/g, Mi).replace(/punctSpace/g, ji).replace(/punct/g, Ai).getRegex(), Vi = q("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Di).replace(/punctSpace/g, Ei).replace(/punct/g, Ti).getRegex(), Hi = q("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Di).replace(/punctSpace/g, Ei).replace(/punct/g, Ti).getRegex(), Ui = q("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Di).replace(/punctSpace/g, Ei).replace(/punct/g, Ti).getRegex(), Wi = q(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, Ti).getRegex(), Gi = q("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Di).replace(/punctSpace/g, Ei).replace(/punct/g, Ti).getRegex(), Ki = q(/\\(punct)/, "gu").replace(/punct/g, Ti).getRegex(), qi = q(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, Ti).getRegex(), Ji = q(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), Yi = q(fi).replace("(?:-->|$)", "-->").getRegex(), Xi = q("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", Yi).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Zi = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Qi = q(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Zi).getRegex(), $i = q(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Qi).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), ea = q(/^!?\[(label)\]\[(ref)\]/).replace("label", Qi).replace("ref", ci).getRegex(), ta = q(/^!?\[(ref)\](?:\[\])?/).replace("ref", ci).getRegex(), na = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, ra = q(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Zi).getRegex(), ia = q("reflink|nolink(?!\\()", "g").replace("reflink", q(/^!?\[(label)\]\[(ref)\]/).replace("label", ra).replace("ref", na).getRegex()).replace("nolink", q(/^!?\[(ref)\](?:\[\])?/).replace("ref", na).getRegex()).getRegex(), aa = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, oa = q(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), sa = {
	_backpedal: Jr,
	anyPunctuation: Ki,
	anyPunctuationOutsideCode: qi,
	autolink: Ji,
	blockSkip: Ni,
	br: Ci,
	code: Si,
	del: Jr,
	delLDelim: Jr,
	delRDelim: Jr,
	emStrongLDelim: Fi,
	emStrongRDelimAst: zi,
	emStrongRDelimUnd: Hi,
	escape: xi,
	link: $i,
	nolink: ta,
	punctuation: Oi,
	reflink: ea,
	reflinkSearch: ia,
	tag: Xi,
	text: wi,
	url: Jr
}, ca = {
	...sa,
	emStrongLDelim: Li,
	emStrongRDelimAst: Vi,
	emStrongRDelimUnd: Ui,
	link: q(/^!?\[(label)\]\((.*?)\)/).replace("label", Qi).getRegex(),
	reflink: q(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Qi).getRegex()
}, la = {
	...sa,
	emStrongRDelimAst: Bi,
	emStrongLDelim: Ii,
	delLDelim: Wi,
	delRDelim: Gi,
	url: q(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", oa).replace("protocol", aa).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: q(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", aa).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", Xr ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, ua = {
	...la,
	br: q(Ci).replace("{2,}", "*").getRegex(),
	text: q(la.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, da = {
	normal: _i,
	gfm: yi,
	pedantic: bi
}, fa = {
	normal: sa,
	gfm: la,
	breaks: ua,
	pedantic: ca
}, pa = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, ma = (e) => pa[e];
function ha(e, t) {
	if (t) {
		if (J.escapeTest.test(e)) return e.replace(J.escapeReplace, ma);
	} else if (J.escapeTestNoEncode.test(e)) return e.replace(J.escapeReplaceNoEncode, ma);
	return e;
}
function ga(e) {
	return J.numericCharacterReference.test(e) ? e.replace(J.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function _a(e) {
	try {
		e = encodeURI(e).replace(J.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function va(e, t) {
	let n = e.replace(J.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(J.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(J.slashPipe, "|");
	return n;
}
function ya(e, t, n) {
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
function ba(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && J.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function xa(e) {
	return e.trim().replace(J.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Sa(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Ca = /* @__PURE__ */ new WeakMap();
function wa(e) {
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
function Ta(e, t) {
	let n = wa(t), r = Ca.get(e);
	r ? r.push(n) : Ca.set(e, [n]);
}
function Ea(e) {
	let t = Ca.get(e);
	t && (t.pop(), t.length === 0 && Ca.delete(e));
}
function Da(e, t, n) {
	let r = Ca.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : wa(t).lastPossibleEmStrongCloserIndex[n];
}
function Oa(e, t) {
	let n = Ca.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : wa(t).lastPossibleDelCloserIndex;
}
function ka(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function Aa(e, t, n, r, i) {
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
function ja(e, t, n) {
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
function Ma(e, t, n, r) {
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
var Na = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || Kr;
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
			let e = this.options.pedantic ? t[0] : ba(t[0]);
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
			let e = t[0], n = ja(e, t[3] || "", this.rules);
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
				let t = ya(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: ya(t[0], "\n"),
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
			raw: ya(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = ya(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + ka(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? ka(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => ka(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => ka(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = ba(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && Sa(t[2], "()") !== -1) return;
			let e = xa(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: ya(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = va(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: ya(t[0], "\n"),
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
			for (let e of i) a.rows.push(va(e, a.header.length).map((e, t) => ({
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
				raw: ya(t[0], "\n"),
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
			if (!this.options.pedantic && Ma(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = ya(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Sa(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = ya(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), Aa(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && Ma(e, n[1], r, this.rules)) return;
			let i = t[xa(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return Aa(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Da(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Oa(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : ga(t[0]),
				escaped: e
			};
		}
	}
}, Pa = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || Kr, this.options.tokenizer = this.options.tokenizer || new Na(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: J,
			block: da.normal,
			inline: fa.normal
		};
		this.options.pedantic ? (t.block = da.pedantic, t.inline = fa.pedantic) : this.options.gfm && (t.block = da.gfm, t.inline = this.options.breaks ? fa.breaks : fa.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: da,
			inline: fa
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(J.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(J.tabCharGlobal, "    ").replace(J.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, xa(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== Na.prototype.emStrong || this.tokenizer.del !== Na.prototype.del ? this.#n(e) : e;
		Ta(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Ea(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, xa(t.slice(r + 1, -1)))) return t;
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
}, Fa = class {
	options;
	parser;
	constructor(e) {
		this.options = e || Kr;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(J.notSpaceStart)?.[0], i = e ? e.replace(J.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + ha(r) + "\">" + (n ? i : ha(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : ha(i, !0)) + "</code></pre>\n";
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
		return `<code>${ha(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? ha(n, !0) : this.parser.parseInline(r), o = _a(e);
		if (o === null) return a;
		e = ha(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + ha(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = _a(e);
		if (i === null) return ha(n);
		e = i;
		let a = `<img src="${ha(e)}" alt="${ha(n)}"`;
		return t && (a += ` title="${ha(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : ha(e.text);
	}
}, Ia = class {
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
}, La = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || Kr, this.options.renderer = this.options.renderer || new Fa(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Ia(), this.textRenderer.parser = this;
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
}, Ra = class {
	options;
	block;
	constructor(e) {
		this.options = e || Kr;
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
		return e ? Pa.lex : Pa.lexInline;
	}
	provideParser(e = this.block) {
		return e ? La.parse : La.parseInline;
	}
};
function za(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var Ba = new class {
	defaults = Gr();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = La;
	Renderer = Fa;
	TextRenderer = Ia;
	Lexer = Pa;
	Tokenizer = Na;
	Hooks = Ra;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (za(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) za(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) za(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				za(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					za(n, this.walkTokens(i, t));
				}) : e.tokens && za(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new Fa(this.defaults);
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
				let t = this.defaults.tokenizer || new Na(this.defaults);
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
				let t = this.defaults.hooks || new Ra();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = Ra.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && Ra.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return za(n, r.call(this, e)), t && za(n, t.call(this, e)), n;
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
		return Pa.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return La.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? Pa.lex : Pa.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? La.parse : La.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? Pa.lex : Pa.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? La.parse : La.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + ha(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function Y(e, t) {
	return Ba.parse(e, t);
}
Y.options = Y.setOptions = function(e) {
	return Ba.setOptions(e), Y.defaults = Ba.defaults, qr(Y.defaults), Y;
}, Y.getDefaults = Gr, Y.defaults = Kr;
function Va(...e) {
	return Ba.use(...e), Y.defaults = Ba.defaults, qr(Y.defaults), Y;
}
Y.use = Va, Y.walkTokens = function(e, t) {
	return Ba.walkTokens(e, t);
}, Y.parseInline = Ba.parseInline, Y.Parser = La, Y.parser = La.parse, Y.Renderer = Fa, Y.TextRenderer = Ia, Y.Lexer = Pa, Y.lexer = Pa.lex, Y.Tokenizer = Na, Y.Hooks = Ra, Y.parse = Y, Y.options, Y.setOptions, Y.walkTokens, Y.parseInline, La.parse, Pa.lex;
// node_modules/dompurify/dist/purify.es.mjs
function Ha(e, t) {
	this.v = e, this.k = t;
}
function Ua(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function Wa(e) {
	if (Array.isArray(e)) return e;
}
function Ga(e, t) {
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
function Ka() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function qa(e, t) {
	return Wa(e) || Ga(e, t) || Ja(e, t) || Ka();
}
function Ja(e, t) {
	if (e) {
		if (typeof e == "string") return Ua(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? Ua(e, t) : void 0;
	}
}
function Ya(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof Ha;
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
Ya.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, Ya.prototype.next = function(e) {
	return this._invoke("next", e);
}, Ya.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, Ya.prototype.return = function(e) {
	return this._invoke("return", e);
};
var Xa = Object.entries, Za = Object.setPrototypeOf, Qa = Object.isFrozen, $a = Object.getPrototypeOf, eo = Object.getOwnPropertyDescriptor, to = Object.freeze, no = Object.seal, ro = Object.create, io = typeof Reflect < "u" && Reflect, ao = io.apply, oo = io.construct;
to ||= function(e) {
	return e;
}, no ||= function(e) {
	return e;
}, ao ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, oo ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var so = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var co = Z(Array.prototype.lastIndexOf), lo = Z(Array.prototype.pop), uo = Z(Array.prototype.push), fo = Z(Array.prototype.slice), po = Z(Array.prototype.splice), mo = Array.isArray, ho = jo(Set), go = Z(Set.prototype.add), _o = Z(Set.prototype.has), vo = Z(String.prototype.toLowerCase), yo = Z(String.prototype.toString), bo = Z(String.prototype.match), xo = Z(String.prototype.replace), So = Z(String.prototype.indexOf), Co = Z(String.prototype.trim), wo = Z(Number.prototype.toString), To = Z(Boolean.prototype.toString), Eo = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), Do = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), Oo = Z(Object.prototype.hasOwnProperty), ko = Z(Object.prototype.toString), X = Z(RegExp.prototype.test), Ao = jo(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return ao(e, t, n);
	};
}
function jo(e) {
	return function() {
		return oo(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : vo;
	if (Za && Za(e, null), !mo(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Qa(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function Mo(e) {
	for (let t = 0; t < e.length; t++) Oo(e, t) || (e[t] = null);
	return e;
}
function No(e) {
	let t = ro(null);
	for (let r of Xa(e)) {
		var n = qa(r, 2);
		let i = n[0], a = n[1];
		Oo(e, i) && (t[i] = mo(a) ? Mo(a) : a && typeof a == "object" && a.constructor === Object ? No(a) : a);
	}
	return t;
}
function Po(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return wo(e);
		case "boolean": return To(e);
		case "bigint": return Eo ? Eo(e) : "0";
		case "symbol": return Do ? Do(e) : "Symbol()";
		case "undefined": return ko(e);
		case "function":
		case "object": {
			if (e === null) return ko(e);
			let t = e, n = Fo(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : ko(e);
			}
			return ko(e);
		}
		default: return ko(e);
	}
}
function Fo(e, t) {
	for (; e !== null;) {
		let n = eo(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = $a(e);
	}
	function n() {
		return null;
	}
	return n;
}
function Io(e, t) {
	for (; e !== null;) {
		let n = eo(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = $a(e);
	}
	return null;
}
function Lo(e) {
	try {
		return X(e, ""), !0;
	} catch {
		return !1;
	}
}
var Ro = to(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), zo = to(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), Bo = to([
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
]), Vo = to([
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
]), Ho = to(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), Uo = to([
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
]), Wo = to(["#text"]), Go = to(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), Ko = to(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), qo = to(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), Jo = to([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), Yo = no(/{{[\w\W]*|^[\w\W]*}}/g), Xo = no(/<%[\w\W]*|^[\w\W]*%>/g), Zo = no(/\${[\w\W]*/g), Qo = no(/^data-[\-\w.\u00B7-\uFFFF]+$/), $o = no(/^aria-[\-\w]+$/), es = no(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), ts = no(/^(?:\w+script|data):/i), ns = no(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), rs = no(/^html$/i), is = no(/^[a-z][.\w]*(-[.\w]+)+$/i), as = no(/<[/\w!]/g), os = no(/<[/\w]/g), ss = no(/<\/no(script|embed|frames)/i), cs = no(/\/>/i), $ = {
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
}, ls = to(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), us = function() {
	return typeof window > "u" ? null : window;
}, ds = function(e, t) {
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
}, fs = function() {
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
}, ps = function(e, t, n, r) {
	return Oo(e, t) && mo(e[t]) ? Q(r.base ? No(r.base) : {}, e[t], r.transform) : n;
}, ms = function(e, t, n) {
	let r = Oo(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? No(r) : n();
};
function hs() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : us(), t = (e) => hs(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = Fo(d, "cloneNode"), p = Fo(d, "remove"), m = Fo(d, "removeAttributeNode"), h = Fo(d, "nextSibling"), g = Fo(d, "childNodes"), _ = Fo(d, "parentNode"), v = Fo(d, "shadowRoot"), y = Fo(d, "attributes"), b = o && o.prototype ? Fo(o.prototype, "nodeType") : null, x = o && o.prototype ? Fo(o.prototype, "nodeName") : null, S = o && o.prototype ? Fo(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, ee = function(e) {
		return x ? x(e) : e.nodeName;
	}, te = Io(d, "namespaceURI"), ne = function(e) {
		return te ? te(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let w, re = "", ie, ae = !1, oe = 0, se = function() {
		if (oe > 0) throw Ao("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
	}, ce = function(e) {
		se(), oe++;
		try {
			return w.createHTML(e);
		} finally {
			oe--;
		}
	}, le = function(e) {
		se(), oe++;
		try {
			return w.createScriptURL(e);
		} finally {
			oe--;
		}
	}, ue = function() {
		return ae ||= (ie = ds(u, i), !0), ie;
	}, de = n, fe = de.implementation, pe = de.createNodeIterator, me = de.createDocumentFragment, he = de.getElementsByTagName, ge = r.importNode, T = fs();
	t.isSupported = typeof Xa == "function" && typeof _ == "function" && fe && fe.createHTMLDocument !== void 0;
	let _e = Yo, E = Xo, ve = Zo, D = Qo, ye = $o, be = ts, O = ns, xe = is, Se = es, k = null, Ce = Q({}, [
		...Ro,
		...zo,
		...Bo,
		...Ho,
		...Wo
	]), A = null, we = Q({}, [
		...Go,
		...Ko,
		...qo,
		...Jo
	]), Te = Object.seal(ro(null, {
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
	})), Ee = null, De = null, Oe = Object.seal(ro(null, {
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
	})), ke = !0, Ae = !0, je = !1, Me = !0, Ne = !1, Pe = !0, Fe = !1, Ie = !1, Le = null, Re = null, ze = !1, Be = !1, Ve = !1, He = !1, Ue = !0, We = !1, Ge = "user-content-", j = !0, Ke = !1, qe = null, Je = {}, Ye = null, Xe = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Ze = null, Qe = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), $e = null, et = Q({}, [
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
	]), tt = "http://www.w3.org/1998/Math/MathML", nt = "http://www.w3.org/2000/svg", rt = "http://www.w3.org/1999/xhtml", it = rt, at = !1, ot = null, st = Q({}, [
		tt,
		nt,
		rt
	], yo), ct = to([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), lt = Q({}, ct), M = to(["annotation-xml"]), ut = Q({}, M), dt = Q({}, [
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
		e && typeof e == "object" || (e = {}), e = No(e), ft = pt.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, N = ft === "application/xhtml+xml" ? yo : vo, k = ps(e, "ALLOWED_TAGS", Ce, { transform: N }), A = ps(e, "ALLOWED_ATTR", we, { transform: N }), ot = ps(e, "ALLOWED_NAMESPACES", st, { transform: yo }), $e = ps(e, "ADD_URI_SAFE_ATTR", et, {
			transform: N,
			base: et
		}), Ze = ps(e, "ADD_DATA_URI_TAGS", Qe, {
			transform: N,
			base: Qe
		}), Ye = ps(e, "FORBID_CONTENTS", Xe, { transform: N }), Ee = ps(e, "FORBID_TAGS", No({}), { transform: N }), De = ps(e, "FORBID_ATTR", No({}), { transform: N }), Je = Oo(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? No(e.USE_PROFILES) : e.USE_PROFILES : !1, ke = e.ALLOW_ARIA_ATTR !== !1, Ae = e.ALLOW_DATA_ATTR !== !1, je = e.ALLOW_UNKNOWN_PROTOCOLS || !1, Me = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Ne = e.SAFE_FOR_TEMPLATES || !1, Pe = e.SAFE_FOR_XML !== !1, Fe = e.WHOLE_DOCUMENT || !1, Be = e.RETURN_DOM || !1, Ve = e.RETURN_DOM_FRAGMENT || !1, He = e.RETURN_TRUSTED_TYPE || !1, ze = e.FORCE_BODY || !1, Ue = e.SANITIZE_DOM !== !1, We = e.SANITIZE_NAMED_PROPS || !1, j = e.KEEP_CONTENT !== !1, Ke = e.IN_PLACE || !1, Se = Lo(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : es, it = typeof e.NAMESPACE == "string" ? e.NAMESPACE : rt, lt = ms(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, ct)), ut = ms(e, "HTML_INTEGRATION_POINTS", () => Q({}, M));
		let t = ms(e, "CUSTOM_ELEMENT_HANDLING", () => ro(null));
		if (Te = ro(null), Oo(t, "tagNameCheck") && gt(t.tagNameCheck) && (Te.tagNameCheck = t.tagNameCheck), Oo(t, "attributeNameCheck") && gt(t.attributeNameCheck) && (Te.attributeNameCheck = t.attributeNameCheck), Oo(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (Te.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), no(Te), Ne && (Ae = !1), Ve && (Be = !0), Je && (k = Q({}, Wo), A = ro(null), Je.html === !0 && (Q(k, Ro), Q(A, Go)), Je.svg === !0 && (Q(k, zo), Q(A, Ko), Q(A, Jo)), Je.svgFilters === !0 && (Q(k, Bo), Q(A, Ko), Q(A, Jo)), Je.mathMl === !0 && (Q(k, Ho), Q(A, qo), Q(A, Jo))), Oe.tagCheck = null, Oe.attributeCheck = null, Oo(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? Oe.tagCheck = e.ADD_TAGS : mo(e.ADD_TAGS) && (k === Ce && (k = No(k)), Q(k, e.ADD_TAGS, N))), Oo(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? Oe.attributeCheck = e.ADD_ATTR : mo(e.ADD_ATTR) && (A === we && (A = No(A)), Q(A, e.ADD_ATTR, N))), Oo(e, "ADD_FORBID_CONTENTS") && mo(e.ADD_FORBID_CONTENTS) && (Ye === Xe && (Ye = No(Ye)), Q(Ye, e.ADD_FORBID_CONTENTS, N)), j && (k["#text"] = !0), Fe && Q(k, [
			"html",
			"head",
			"body"
		]), k.table && (Q(k, ["tbody"]), delete Ee.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw Ao("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw Ao("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = w;
			w = e.TRUSTED_TYPES_POLICY;
			try {
				re = ce("");
			} catch (e) {
				throw w = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (w = void 0, re = "") : (w === void 0 && (w = ue()), w && typeof re == "string" && (re = ce("")));
		to && to(e), mt = e;
	}, vt = Q({}, [
		...zo,
		...Bo,
		...Vo
	]), yt = Q({}, [...Ho, ...Uo]), bt = function(e, t, n) {
		return t.namespaceURI === rt ? e === "svg" : t.namespaceURI === tt ? e === "svg" && (n === "annotation-xml" || lt[n]) : !!vt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === rt ? e === "math" : t.namespaceURI === nt ? e === "math" && ut[n] : !!yt[e];
	}, St = function(e, t, n) {
		return t.namespaceURI === nt && !ut[n] || t.namespaceURI === tt && !lt[n] ? !1 : !yt[e] && (dt[e] || !vt[e]);
	}, Ct = function(e) {
		let t = ne(e), n = vo(ee(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: ne(r) }, a = vo(ee(r))) : (i = { namespaceURI: it }, a = "template"), ot[t] ? t === nt ? bt(n, i, a) : t === tt ? xt(n, i, a) : t === rt ? St(n, i, a) : !!(ft === "application/xhtml+xml" && ot[t]) : !1;
	}, wt = function(e) {
		uo(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw Ao("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, Tt = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, Et = function(e) {
		At(e);
		let t = g(e);
		t && so(fo(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && Tt(e, r, i);
		}
	}, Dt = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		uo(t.removed, {
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
			if (Be || Ve) try {
				wt(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Ot = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || A[N(i)] || Tt(e, r, i);
		}
	}, kt = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				P(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, At = function(e) {
		kt(e, (e) => {
			C(e) === $.element && Ot(e);
		});
	}, jt = function(e) {
		let t = [];
		return kt(e, (e) => {
			t.push(e);
		}, !0), t;
	}, Mt = function(e, t) {
		let n = ho();
		so(jt(e), (e) => {
			C(e) === $.element && go(n, e);
		}), so(t, (e) => {
			C(e) === $.element && !_o(n, e) && Ot(e);
		});
	}, Nt = function(e, n) {
		let r = !1;
		return so(t.removed, (t) => {
			t.element && (r ||= t.element === e, At(t.element));
		}), Mt(e, n), r;
	}, Pt = function(e, t) {
		let n = qe, r = null, i = () => (r || (r = ho(), so(t, (e) => {
			go(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = jt(e), a = [e], o = !1;
			if (so(r, (e) => {
				C(e) === $.documentFragment ? uo(a, e) : !o && !_o(n, e) && (o = _o(i(), e));
			}), !o) return;
			if (t >= 8) throw Ao("the tree kept changing during in-place sanitization; refusing to return it");
			so(a, (t, n) => {
				Ft(t, n > 0, e, i());
			});
		}
	}, Ft = function(e, t, n, r) {
		let i = Bt(e), a;
		for (; a = i.nextNode();) _o(r, a) && !_o(qe, a) && (t && F(T.uponSanitizeShadowNode, a, null), Jt(a, n), tn(a, n), P(a.content) && nn(a.content));
	}, It = function(e) {
		qe && go(qe, e);
	}, Lt = function(e, t) {
		return Pe ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Rt = function(e) {
		Pe && kt(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && X(os, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = N(ee(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Lt("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, zt = function(e) {
		let t = null, r = null;
		if (ze) e = "<remove></remove>" + e;
		else {
			let t = bo(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		ft === "application/xhtml+xml" && it === rt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = w ? ce(e) : e;
		if (it === rt) try {
			t = new l().parseFromString(i, ft);
		} catch {}
		if (!t || !t.documentElement) {
			t = fe.createDocument(it, "template", null);
			try {
				t.documentElement.innerHTML = at ? re : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), it === rt ? he.call(t, Fe ? "html" : "body")[0] : Fe ? t.documentElement : a;
	}, Bt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return pe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Vt = function(e) {
		return e = xo(e, _e, " "), e = xo(e, E, " "), e = xo(e, ve, " "), e;
	}, Ht = function(e) {
		Ut(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && so(t, (e) => {
			P(e.content) && Ht(e.content);
		});
	}, Ut = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = pe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Vt(r.data), r = n.nextNode();
	}, Wt = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || N(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, P = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, Gt = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function F(e, n, r) {
		e.length !== 0 && so(e, (e) => {
			e.call(t, n, r, mt);
		});
	}
	let I = function(e, t) {
		return !!(Pe && e.hasChildNodes() && !Gt(e.firstElementChild) && X(as, e.textContent) && X(as, e.innerHTML) || Pe && ls[t] && C(e) === $.element && ne(e) === rt && (Gt(e.firstElementChild) || typeof e.textContent == "string" && X(as, e.textContent)) || e.nodeType === $.processingInstruction || Pe && e.nodeType === $.comment && X(os, e.data));
	}, Kt = function(e, t) {
		return e instanceof RegExp ? X(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, L = function(e, t, n) {
		if (!Ee[t] && Zt(t) && Kt(Te.tagNameCheck, t)) return !1;
		if (j && !Ye[t] && !Ye[vo(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return wt(e), !0;
	}, R = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? No(t) : t;
	}, qt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ke && At(e), !0);
	}, Jt = function(e, n) {
		if (It(e), F(T.beforeSanitizeElements, e, null), qt(e, n)) return !0;
		if (Wt(e)) return wt(e), !0;
		let r = N(ee(e));
		if (k = R(T.uponSanitizeElement, k, Ce, Le), F(T.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: k
		}), qt(e, n)) return !0;
		if (I(e, r)) return wt(e), !0;
		if (Ee[r] || !(Oe.tagCheck instanceof Function && Oe.tagCheck(r)) && !k[r]) {
			let t = L(e, r, n);
			return t === !1 && (F(T.afterSanitizeElements, e, null), qt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !Ct(e) || (r === "noscript" || r === "noembed" || r === "noframes") && X(ss, e.innerHTML)) return wt(e), !0;
		if (Ne && e.nodeType === $.text) {
			let n = Vt(e.textContent);
			e.textContent !== n && (uo(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return F(T.afterSanitizeElements, e, null), qt(e, n);
	}, Yt = function(e, t, r) {
		if (De[t] || Lt(t, e) || Ue && (t === "id" || t === "name") && (r in n || r in ht)) return !1;
		let i = A[t] || Oe.attributeCheck instanceof Function && Oe.attributeCheck(t, e);
		return Ae && X(D, t) || ke && X(ye, t) ? !0 : i ? $e[t] || X(Se, xo(r, O, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && So(r, "data:") === 0 && Ze[e] || je && !X(be, xo(r, O, "")) ? !0 : !r : Zt(e) && Kt(Te.tagNameCheck, e) && Kt(Te.attributeNameCheck, t, e) || t === "is" && Te.allowCustomizedBuiltInElements && Kt(Te.tagNameCheck, r);
	}, Xt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), Zt = function(e) {
		return !Xt[vo(e)] && X(xe, e);
	}, Qt = function(e, t, n, r) {
		if (w && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return ce(r);
			case "TrustedScriptURL": return le(r);
		}
		return r;
	}, $t = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Wt(e) || (wt(e), !1);
		} catch {
			return Dt(t, e), !1;
		}
	}, en = function(e, t) {
		return Pe && X(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : vo(e) === "attributename" && bo(t, "href") !== null;
	}, tn = function(e, n) {
		if (F(T.beforeSanitizeAttributes, e, null), qt(e, n)) return;
		let r = e.attributes;
		if (!r || Wt(e)) return;
		A = R(T.uponSanitizeAttribute, A, we, Re);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: A,
			forceKeepAttr: void 0
		}, a = r.length, o = N(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = N(s), d = l, f = s === "value" ? d : Co(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, F(T.uponSanitizeAttribute, e, i), f = i.attrValue, We && (u === "id" || u === "name") && So(f, Ge) !== 0 && (Dt(s, e, n), f = Ge + f, p = !0), en(s, f) ? Dt(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !Me && X(cs, f) ? Dt(s, e, n) : (Ne && (f = Vt(f)), Yt(o, u, f) ? (f = Qt(o, u, c, f), f !== d && $t(e, s, c, f) && p && lo(t.removed)) : Dt(s, e, n)) : Dt(s, e, n));
		}
		F(T.afterSanitizeAttributes, e, null), qt(e, n);
	}, nn = function(e) {
		let t = null, n = Bt(e);
		for (F(T.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (F(T.uponSanitizeShadowNode, t, null), Jt(t, e), tn(t, e), P(t.content) && nn(t.content), C(t) === $.element) {
			let e = v(t);
			P(e) && (rn(e), nn(e));
		}
		Ne && Ut(e), F(T.afterSanitizeShadowDOM, e, null);
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
				if (typeof e == "string" && N(e) === "template") {
					let e = n.content;
					P(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				P(e) && t.push({
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
		if (at = !e, at && (e = "<!-->"), typeof e != "string" && !Gt(e) && (e = Po(e), typeof e != "string")) throw Ao("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Ie ? (k = Le, A = Re) : _t(n), (T.uponSanitizeElement.length > 0 || T.uponSanitizeAttribute.length > 0) && (k = No(k)), T.uponSanitizeAttribute.length > 0 && (A = No(A)), t.removed = [];
		let c = Ke && typeof e != "string" && Gt(e), l = [], u = qe;
		if (c) {
			Rt(e);
			let t = ee(e);
			if (typeof t == "string") {
				let n = N(t);
				if (!k[n] || Ee[n]) throw Et(e), Ao("root node is forbidden and cannot be sanitized in-place");
			}
			if (Wt(e)) throw Et(e), Ao("root node is clobbered and cannot be sanitized in-place");
			l = jt(e), qe = ho();
		} else if (Gt(e)) i = zt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), rn(i);
		else {
			if (!Be && !Ne && !Fe && e.indexOf("<") === -1) return w && He ? ce(e) : e;
			if (i = zt(e), !i) return Be ? null : He ? re : "";
		}
		i && ze && wt(i.firstChild);
		let d = c ? e : i;
		try {
			c && rn(e);
			let t = Bt(d);
			for (; o = t.nextNode();) Jt(o, d), tn(o, d), P(o.content) && nn(o.content);
			c && Pt(e, l);
		} catch (t) {
			throw c && (Et(e), Nt(e, l)), t;
		} finally {
			qe = u;
		}
		if (c) {
			if (Nt(e, l)) throw Ao("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Ne && Ht(e), e;
		}
		if (Be) {
			if (Ne && Ht(i), Ve) for (s = me.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (A.shadowroot || A.shadowrootmode) && (s = ge.call(r, s, !0)), s;
		}
		let f = Fe ? i.outerHTML : i.innerHTML;
		return Fe && k["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && X(rs, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Ne && (f = Vt(f)), w && He ? ce(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		_t(e), Ie = !0, Le = k, Re = A;
	}, t.clearConfig = function() {
		mt = null, Ie = !1, Le = null, Re = null, w = ie, re = "";
	}, t.isValidAttribute = function(e, t, n) {
		mt || _t({});
		let r = N(e), i = N(t);
		return typeof n == "string" && (en(t, n) || !Me && X(cs, n)) ? !1 : Yt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && Oo(T, e) && uo(T[e], t);
	}, t.removeHook = function(e, t) {
		if (Oo(T, e)) {
			if (t !== void 0) {
				let n = co(T[e], t);
				return n === -1 ? void 0 : po(T[e], n, 1)[0];
			}
			return lo(T[e]);
		}
	}, t.removeHooks = function(e) {
		Oo(T, e) && (T[e] = []);
	}, t.removeAllHooks = function() {
		T = fs();
	}, t;
}
hs();
// kit/format.js
function gs(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function _s(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
function vs(e) {
	return !e || !e.meta ? {} : typeof e.meta == "string" ? JSON.parse(e.meta) : e.meta;
}
function ys(e) {
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
function bs(e) {
	let t = vs(e), n = [], r = t.github || {}, i = r.repo || "elee1766/jobshop";
	r.issue && n.push({
		label: "issue #" + r.issue,
		href: `https://github.com/${i}/issues/${r.issue}`
	}), r.pr && n.push({
		label: "PR #" + r.pr,
		href: `https://github.com/${i}/pull/${r.pr}`
	});
	let a = t.linear || {}, o = a.issue || a.identifier, s = a.url || (a.workspace && o ? `https://linear.app/${encodeURIComponent(a.workspace)}/issue/${encodeURIComponent(o)}` : "");
	return s && xs(s) && n.push({
		label: o || "Linear",
		href: s
	}), n;
}
function xs(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
function Ss(e) {
	return !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e);
}
// kit/Holder.svelte
var Cs = /* @__PURE__ */ U("<span class=\"svelte-bwb1ob\"> </span>"), ws = /* @__PURE__ */ U("<span class=\"faint svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), Ts = /* @__PURE__ */ U("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), Es = /* @__PURE__ */ U("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span> <!> <!></button>");
function Ds(e, t) {
	Re(t, !0);
	let n = Hr(t, "label", 3, "");
	var r = Es(), i = I(r);
	let a;
	var o = R(i, 2), s = (e) => {
		var t = Cs(), r = L(t, !0);
		z(() => G(r, n())), W(e, t);
	};
	K(o, (e) => {
		n() && e(s);
	});
	var c = R(o, 2), l = L(c, !0), u = R(c, 2), d = (e) => {
		var n = ws(), r = L(n, !0);
		z((e) => G(r, e), [() => [t.h.pool, t.h.model].filter(Boolean).join(" · ")]), W(e, n);
	};
	K(u, (e) => {
		(t.h.pool || t.h.model) && e(d);
	});
	var f = R(u, 2), p = (e) => {
		var n = Ts(), r = L(n, !0);
		z(() => G(r, t.h.tool || "working")), W(e, n);
	};
	K(f, (e) => {
		t.h.busy && e(p);
	}), O(r), z(() => {
		Ir(r, "title", "Open " + t.h.id + " read-only"), Ir(r, "data-session", t.h.id), a = Mr(i, 1, "dot svelte-bwb1ob", null, a, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		}), G(l, t.h.display_name || t.h.id);
	}), $n("click", r, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), W(e, r), ze();
}
er(["click"]);
// kit/Trace.svelte
var Os = /* @__PURE__ */ U("<span class=\"faint\"> </span>"), ks = /* @__PURE__ */ U("<span data-testid=\"trace-session\"><!></span>"), As = /* @__PURE__ */ U("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"faint svelte-vk0dlq\"> </a>"), js = /* @__PURE__ */ U("<div class=\"kids svelte-vk0dlq\"></div>"), Ms = /* @__PURE__ */ U("<div class=\"node svelte-vk0dlq\"><a class=\"mono svelte-vk0dlq\"> </a> <span> </span> <span class=\"t svelte-vk0dlq\"> </span> <!> <!> <!> <!></div>");
function Ns(e, t) {
	Re(t, !0);
	var n = Ms(), r = I(n), i = L(r, !0), a = R(r, 2), o = L(a, !0), s = R(a, 2), c = L(s, !0), l = R(s, 2), u = (e) => {
		var n = Os(), r = L(n);
		z(() => G(r, `via ${t.node.trigger ?? ""}`)), W(e, n);
	};
	K(l, (e) => {
		t.node.trigger && e(u);
	});
	var d = R(l, 2), f = (e) => {
		var n = ks(), r = I(n);
		{
			let e = /* @__PURE__ */ nt(() => t.node.session ?? { id: t.node.holder });
			Ds(r, {
				get h() {
					return H(e);
				},
				get onopen() {
					return t.watch;
				},
				label: "View session"
			});
		}
		O(n), W(e, n);
	};
	K(d, (e) => {
		t.node.holder && t.watch && e(f);
	});
	var p = R(d, 2);
	wr(p, 16, () => t.node.links ?? [], (e) => e, (e, t) => {
		var n = cr(), r = Kt(n), i = (e) => {
			var n = As(), r = L(n, !0);
			z((e) => {
				Ir(n, "href", t), G(r, e);
			}, [() => t.replace(/^https:\/\/github\.com\//, "")]), W(e, n);
		}, a = /* @__PURE__ */ nt(() => xs(t));
		K(r, (e) => {
			H(a) && e(i);
		}), W(e, n);
	});
	var m = R(p, 2), h = (e) => {
		var n = js();
		wr(n, 21, () => t.node.children, (e) => e.id, (e, n) => {
			Ns(e, {
				get node() {
					return H(n);
				},
				get go() {
					return t.go;
				},
				get watch() {
					return t.watch;
				}
			});
		}), O(n), W(e, n);
	};
	K(m, (e) => {
		t.node.children?.length && e(h);
	}), O(n), z((e) => {
		Ir(r, "href", "?ticket=" + t.node.id), G(i, t.node.id), Mr(a, 1, `pill ${e ?? ""}`, "svelte-vk0dlq"), G(o, t.node.status), G(c, t.node.title);
	}, [() => ys(t.node.status)]), $n("click", r, (e) => {
		e.preventDefault(), t.go({ ticket: t.node.id });
	}), W(e, n), ze();
}
er(["click"]);
// plugins/ticket/Details.svelte
var Ps = /* @__PURE__ */ U("<a class=\"mono sl svelte-jqeuhg\"> </a>"), Fs = /* @__PURE__ */ U("<span class=\"pill\"> </span>"), Is = /* @__PURE__ */ U("<span class=\"pill accent\"> </span>"), Ls = /* @__PURE__ */ U("<a target=\"_blank\" rel=\"noopener noreferrer\"> </a>"), Rs = /* @__PURE__ */ U("<span class=\"muted held svelte-jqeuhg\" data-testid=\"held-by\">held by <!> <button type=\"button\" class=\"ghost\" data-testid=\"message-session\">Message this session</button></span>"), zs = /* @__PURE__ */ U("<span class=\"mono\"> </span>"), Bs = /* @__PURE__ */ U("<span class=\"muted\">held by <!></span>"), Vs = /* @__PURE__ */ U("<div class=\"row svelte-jqeuhg\"><span class=\"muted\">parent</span> <a class=\"mono\"> </a></div>"), Hs = /* @__PURE__ */ U("<span> </span>"), Us = /* @__PURE__ */ U("<div class=\"report svelte-jqeuhg\"><span class=\"pill accent\"> </span> <!> <span class=\"faint\"> </span></div>"), Ws = /* @__PURE__ */ U("<span class=\"pill bad\"> </span>"), Gs = /* @__PURE__ */ U("<div class=\"muted note svelte-jqeuhg\"> <span class=\"faint\"> </span></div>"), Ks = /* @__PURE__ */ U("<li class=\"svelte-jqeuhg\"><div class=\"stepline svelte-jqeuhg\"><a class=\"svelte-jqeuhg\"> </a> <span> </span> <!> <!></div> <!></li>"), qs = /* @__PURE__ */ U("<h3 class=\"svelte-jqeuhg\">Steps</h3> <ol class=\"steps svelte-jqeuhg\"></ol>", 1), Js = /* @__PURE__ */ U("<h3 class=\"svelte-jqeuhg\">Body</h3> <pre class=\"body svelte-jqeuhg\"> </pre>", 1), Ys = /* @__PURE__ */ U("<h3 class=\"svelte-jqeuhg\">Trace</h3> <!>", 1), Xs = /* @__PURE__ */ U("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"svelte-jqeuhg\">open</a>"), Zs = /* @__PURE__ */ U("<li class=\"svelte-jqeuhg\"><span class=\"mono\"> </span> <span class=\"faint\"> </span> <!></li>"), Qs = /* @__PURE__ */ U("<h3 class=\"svelte-jqeuhg\">External activity</h3> <ul class=\"ext svelte-jqeuhg\" data-testid=\"drawer-activity\"></ul>", 1), $s = /* @__PURE__ */ U("<div class=\"comment svelte-jqeuhg\"><div class=\"faint\"><!> </div> <pre> </pre></div>"), ec = /* @__PURE__ */ U("<h3 class=\"svelte-jqeuhg\">Comments</h3> <!>", 1), tc = /* @__PURE__ */ U("<div class=\"ticket\"><h2 class=\"svelte-jqeuhg\"> </h2> <div class=\"row svelte-jqeuhg\"><span class=\"pill\"> </span> <span> </span> <!> <!></div> <div class=\"row svelte-jqeuhg\"><!> <!> <span class=\"faint\"> </span></div> <!> <!> <!> <!> <!> <!> <!></div>");
function nc(e, t) {
	Re(t, !0);
	let n = (e, t = m) => {
		var n = Ps(), r = L(n, !0);
		z(() => {
			Ir(n, "href", "?session=" + t()), G(r, t());
		}), $n("click", n, (e) => {
			e.preventDefault(), e.stopPropagation(), s(t());
		}), W(e, n);
	}, r = Hr(t, "context", 19, () => ({})), i = (e) => t.api.go(e), a = /* @__PURE__ */ nt(() => r().ticket), o = Wr(() => t.api, () => H(a) ? "api/activity?limit=20&ticket=" + encodeURIComponent(H(a).id) : null), s = (e) => i({
		session: e,
		ticket: ""
	}), c = (e) => t.api.openSession(e, { readonly: !0 }), l = (e) => r().sessions?.[e] ?? { id: e }, u = (e) => vs(e).report || null;
	var d = cr(), f = Kt(d), p = (e) => {
		let t = /* @__PURE__ */ nt(() => _s(H(a))), d = /* @__PURE__ */ nt(() => u(H(a)));
		var f = tc(), p = I(f), m = L(p, !0), h = R(p, 2), g = I(h), _ = L(g, !0), v = R(g, 2), y = L(v), b = R(v, 2), x = (e) => {
			var t = Fs(), n = L(t, !0);
			z(() => G(n, H(a).pool)), W(e, t);
		};
		K(b, (e) => {
			H(a).pool && e(x);
		}), wr(R(b, 2), 16, () => H(a).labels ?? [], (e) => e, (e, t) => {
			var n = Is(), r = L(n, !0);
			z(() => G(r, t)), W(e, n);
		}), O(h);
		var S = R(h, 2), C = I(S);
		wr(C, 17, () => bs(H(a)), (e) => e.href, (e, t) => {
			var n = Ls(), r = L(n, !0);
			z(() => {
				Ir(n, "href", H(t).href), G(r, H(t).label);
			}), W(e, n);
		});
		var ee = R(C, 2), te = (e) => {
			var t = cr(), r = Kt(t), i = (e) => {
				var t = Rs(), n = R(I(t));
				{
					let e = /* @__PURE__ */ nt(() => l(H(a).holder));
					Ds(n, {
						get h() {
							return H(e);
						},
						onopen: c,
						label: "View session"
					});
				}
				var r = R(n, 2);
				O(t), $n("click", r, () => s(H(a).holder)), W(e, t);
			}, o = /* @__PURE__ */ nt(() => Ss(H(a).holder) && H(a).status === "in_progress"), u = (e) => {
				var t = Bs(), r = R(I(t)), i = (e) => {
					n(e, () => H(a).holder);
				}, o = /* @__PURE__ */ nt(() => Ss(H(a).holder)), s = (e) => {
					var t = zs(), n = L(t, !0);
					z(() => G(n, H(a).holder)), W(e, t);
				};
				K(r, (e) => {
					H(o) ? e(i) : e(s, -1);
				}), O(t), W(e, t);
			};
			K(r, (e) => {
				H(o) ? e(i) : e(u, -1);
			}), W(e, t);
		};
		K(ee, (e) => {
			H(a).holder && e(te);
		});
		var ne = L(R(ee, 2));
		O(S);
		var w = R(S, 2), re = (e) => {
			var t = Vs(), n = R(I(t), 2), r = L(n, !0);
			O(t), z(() => {
				Ir(n, "href", "?ticket=" + H(a).parent_id), G(r, H(a).parent_id);
			}), $n("click", n, (e) => {
				e.preventDefault(), i({ ticket: H(a).parent_id });
			}), W(e, t);
		};
		K(w, (e) => {
			H(a).parent_id && e(re);
		});
		var ie = R(w, 2), ae = (e) => {
			var t = Us(), n = I(t), r = L(n, !0), i = R(n, 2), a = (e) => {
				var t = Hs(), n = L(t, !0);
				z(() => G(n, H(d).note)), W(e, t);
			};
			K(i, (e) => {
				H(d).note && e(a);
			});
			var o = L(R(i, 2), !0);
			O(t), z((e) => {
				G(r, H(d).state), G(o, e);
			}, [() => gs(H(d).at)]), W(e, t);
		};
		K(ie, (e) => {
			H(d) && e(ae);
		});
		var oe = R(ie, 2), se = (e) => {
			var t = qs(), n = R(Kt(t), 2);
			wr(n, 21, () => r().run.steps, (e) => e.id, (e, t) => {
				let n = /* @__PURE__ */ nt(() => _s(H(t))), r = /* @__PURE__ */ nt(() => u(H(t)));
				var a = Ks(), o = I(a), s = I(o), d = L(s, !0), f = R(s, 2), p = L(f), m = R(f, 2), h = (e) => {
					var t = Ws(), r = L(t);
					z(() => G(r, `${H(n).failures ?? ""} failures`)), W(e, t);
				};
				K(m, (e) => {
					H(n).failures && e(h);
				});
				var g = R(m, 2), _ = (e) => {
					{
						let n = /* @__PURE__ */ nt(() => l(H(t).holder));
						Ds(e, {
							get h() {
								return H(n);
							},
							onopen: c
						});
					}
				};
				K(g, (e) => {
					H(t).status === "in_progress" && H(t).holder && e(_);
				}), O(o);
				var v = R(o, 2), y = (e) => {
					var t = Gs(), n = I(t), i = L(R(n), !0);
					O(t), z((e) => {
						G(n, `${H(r).state ?? ""}: ${H(r).note ?? ""} `), G(i, e);
					}, [() => gs(H(r).at)]), W(e, t);
				};
				K(v, (e) => {
					H(r)?.note && e(y);
				}), O(a), z((e) => {
					Ir(s, "href", "?ticket=" + H(t).id), G(d, H(n).step || H(t).title), Mr(f, 1, `pill ${e ?? ""}`, "svelte-jqeuhg"), G(p, `${H(t).status ?? ""}${H(t).resolution ? " · " + H(t).resolution : ""}`);
				}, [() => ys(H(t).status)]), $n("click", s, (e) => {
					e.preventDefault(), i({ ticket: H(t).id });
				}), W(e, a);
			}), O(n), W(e, t);
		};
		K(oe, (e) => {
			r().run && e(se);
		});
		var ce = R(oe, 2), le = (e) => {
			var t = Js(), n = L(R(Kt(t), 2), !0);
			z(() => G(n, H(a).body)), W(e, t);
		};
		K(ce, (e) => {
			H(a).body && e(le);
		});
		var ue = R(ce, 2), de = (e) => {
			var t = Ys();
			Ns(R(Kt(t), 2), {
				get node() {
					return r().trace;
				},
				go: i,
				watch: c
			}), W(e, t);
		};
		K(ue, (e) => {
			r().trace && e(de);
		});
		var fe = R(ue, 2), pe = (e) => {
			var t = Qs(), n = R(Kt(t), 2);
			wr(n, 21, () => o.data.items, (e) => e.id, (e, t) => {
				var n = Zs(), r = I(n), i = L(r, !0), a = R(r), o = R(a), s = L(o), c = R(o, 2), l = (e) => {
					var n = Xs();
					z(() => Ir(n, "href", H(t).url)), W(e, n);
				}, u = /* @__PURE__ */ nt(() => H(t).url && xs(H(t).url));
				K(c, (e) => {
					H(u) && e(l);
				}), O(n), z((e) => {
					G(i, H(t).actor), G(a, ` ${H(t).title ?? ""} `), G(s, `· ${e ?? ""}`);
				}, [() => gs(H(t).at)]), W(e, n);
			}), O(n), W(e, t);
		};
		K(fe, (e) => {
			o.data?.items?.length && e(pe);
		});
		var me = R(fe, 2), he = (e) => {
			var t = ec();
			wr(R(Kt(t), 2), 17, () => r().comments, (e) => e.id, (e, t) => {
				var r = $s(), i = I(r), a = I(i), o = (e) => {
					n(e, () => H(t).author);
				}, s = /* @__PURE__ */ nt(() => Ss(H(t).author)), c = (e) => {
					var n = zs(), r = L(n, !0);
					z(() => G(r, H(t).author)), W(e, n);
				};
				K(a, (e) => {
					H(s) ? e(o) : e(c, -1);
				});
				var l = R(a);
				O(i);
				var u = L(R(i, 2), !0);
				O(r), z((e) => {
					G(l, ` · ${e ?? ""}`), G(u, H(t).body);
				}, [() => gs(H(t).created_at)]), W(e, r);
			}), W(e, t);
		};
		K(me, (e) => {
			r().comments?.length && e(he);
		}), O(f), z((e, n) => {
			G(m, H(a).title), G(_, H(a).kind), Mr(v, 1, `pill ${e ?? ""}`, "svelte-jqeuhg"), G(y, `${(H(t).state || H(a).status) ?? ""}${H(a).resolution ? " · " + H(a).resolution : ""}`), G(ne, `updated ${n ?? ""}`);
		}, [() => ys(H(t).state || H(a).status), () => gs(H(a).updated_at)]), W(e, f);
	};
	K(f, (e) => {
		H(a) && e(p);
	}), W(e, d), ze();
}
// plugins/ticket/index.js
er(["click"]), Ur("desk-ticket-details", nc);
