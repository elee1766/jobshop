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
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, C = 65536, te = 1 << 19, ne = 1 << 20, re = 1 << 25, ie = 1 << 21, ae = 1 << 22, oe = 1 << 23, se = Symbol("$state"), ce = Symbol("component"), le = Symbol(""), ue = Symbol("attributes"), de = Symbol("class"), fe = Symbol("style"), pe = Symbol("text"), me = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), he = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function ge() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function _e(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ve() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var w = !1;
function ye(e) {
	w = e;
}
var T;
function E(t) {
	if (t === null) throw _e(), e;
	return T = t;
}
function be() {
	return E(/* @__PURE__ */ I(T));
}
function D(t) {
	if (w) {
		if (/* @__PURE__ */ I(T) !== null) throw _e(), e;
		T = t;
	}
}
function xe(e = 1) {
	if (w) {
		for (var t = e, n = T; t--;) n = /* @__PURE__ */ I(n);
		T = n;
	}
}
function Se(e = !0) {
	for (var t = 0, n = T;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ I(n);
		e && n.remove(), n = i;
	}
}
function Ce(t) {
	if (!t || t.nodeType !== 8) throw _e(), e;
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
var O = null;
function Le(e) {
	O = e;
}
function Re(e, t = !1, n) {
	O = {
		p: O,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: K,
		l: null
	};
}
function ze(e) {
	var t = O, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) tn(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, O = t.p, Be(e);
}
function Be(e = {}) {
	return s(e, ce, { value: !0 }), e;
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
	if (He.length === 0 && !ft) {
		var t = He;
		queueMicrotask(() => {
			t === He && Ue();
		});
	}
	He.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ge = ~(v | y | _);
function k(e, t) {
	e.f = e.f & Ge | t;
}
function Ke(e) {
	e.f & 512 || e.deps === null ? k(e, _) : k(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function qe(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), k(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Je(e) {
	var t = U, n = K;
	G(null), q(null);
	try {
		return e();
	} finally {
		G(t), q(n);
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
	var s = K, c = Xe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				B(e, s);
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
		Promise.all(n.map((e) => /* @__PURE__ */ tt(e))).then(u).catch((e) => B(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Ze());
	}) : f();
}
function Xe() {
	var e = K, t = U, n = O, r = A;
	return function(i = !0) {
		q(e), G(t), Le(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Ze(e = !0) {
	q(null), G(null), Le(null), e && A?.deactivate();
}
function Qe() {
	var e = K, t = e.b, n = A, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function $e(e) {
	var n = 2 | v;
	return K !== null && (K.f |= te), {
		ctx: O,
		deps: null,
		effects: null,
		equals: we,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: K,
		ac: null
	};
}
var et = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function tt(e, n, r) {
	let i = K;
	i === null && De();
	var a = void 0, o = Et(t), s = !U, c = /* @__PURE__ */ new Set();
	return rn(() => {
		var t = K, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== me && n.reject(e);
			}).finally(Ze);
		} catch (e) {
			n.reject(e), Ze();
		}
		var r = A;
		if (s) {
			if (t.f & 32768) var l = Qe();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(et);
			else for (let e of c.values()) e.reject(et);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== et && (r.activate(), t ? (o.f |= oe, jt(o, t)) : (o.f & 8388608 && (o.f ^= oe), jt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), $t(() => {
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
	return xn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function rt(e) {
	let t = /* @__PURE__ */ $e(e);
	return t.equals = Ee, t;
}
function it(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) H(t[n]);
	}
}
function at(e) {
	var n, r = K, i = e.parent;
	if (!yn && i !== null && e.v !== t && i.f & 24576) return ge(), e.v;
	q(i);
	try {
		it(e), n = An(e);
	} finally {
		q(r);
	}
	return n;
}
function ot(e) {
	var t = at(e);
	!e.equals(t) && (e.wv = Dn(), (!A?.is_fork || e.deps === null) && (A === null ? e.v = t : (A.capture(e, t, !0), ut?.capture(e, t, !0)), e.deps === null)) ? k(e, _) : yn || (j === null ? Ke(e) : (Qt() || A?.is_fork) && j.set(e, t));
}
function st(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Je(() => {
		t.ac.abort(me), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Nn(t, 0), ln(t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Pn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var lt = null, A = null, ut = null, j = null, dt = null, ft = !1, pt = !1, mt = null, ht = null, gt = 0, _t = 1, vt = class e {
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
			for (var r of n.d) k(r, v), t(r);
			for (r of n.m) k(r, y), t(r);
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
		for (let [e, t] of this.#s) k(e, t), this.schedule(e);
		this.apply();
		for (var t = mt = [], n = [], r = ht = []; this.#a.length > 0;) {
			gt++ > 1e3 && (this.#b(), yt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Ct(e), this.#d() || this.discard(), t;
			}
		}
		if (A = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (mt = null, ht = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) St(e, t);
			r.length > 0 && A.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), ut = this, bt(n), bt(t), ut = null, this.#i?.resolve();
			var o = A;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (N.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : On(r) && (i & 16 && this.add_dirty_reaction(r, y), Pn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), k(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), A = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) qe(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), j?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		A = this;
	}
	deactivate() {
		A = null, j = null;
	}
	flush() {
		try {
			pt = !0, A = this, this.#m();
		} finally {
			gt = 0, dt = null, mt = null, ht = null, pt = !1, A = null, j = null, N.clear();
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
		if (A === null) {
			let t = A = new e();
			!pt && We(() => {
				t.started || t.flush();
			});
		}
		return A;
	}
	apply() {
		j = null;
	}
	schedule(e) {
		dt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
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
		Me();
	} catch (e) {
		B(e, dt);
	}
}
var M = null;
function bt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && On(r) && (M = /* @__PURE__ */ new Set(), Pn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && fn(r), M?.size > 0)) {
				N.clear();
				for (let e of M) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) M.has(n) && (M.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Pn(n);
					}
				}
				M.clear();
			}
		}
		M = null;
	}
}
function xt(e) {
	A.schedule(e);
}
function St(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), k(e, _);
		for (var n = e.first; n !== null;) St(n, t), n = n.next;
	}
}
function Ct(e) {
	k(e, _);
	for (var t = e.first; t !== null;) Ct(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var wt = /* @__PURE__ */ new Set(), N = /* @__PURE__ */ new Map(), Tt = !1;
function Et(e, t) {
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
function P(e, t) {
	let n = Et(e, t);
	return xn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Dt(e, t = !1, n = !0) {
	let r = Et(e);
	return t || (r.equals = Ee), r;
}
function Ot(e, t, n = !1) {
	return U !== null && (!W || U.f & 131072) && Ve() && U.f & 4325394 && (J === null || !J.has(e)) && Fe(), jt(e, n ? It(t) : t, ht);
}
var kt = null, At = 0;
function jt(e, t, n = null) {
	if (!e.equals(t)) {
		yn ? N.set(e, t) : N.has(e) || N.set(e, e.v);
		var r = vt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && at(t), j === null && Ke(t);
		}
		e.wv = Dn(), kt = null, At = 0, Ft(e, v, n), kt = null, Ve() && K !== null && K.f & 1024 && !(K.f & 96) && (Z === null ? Sn([e]) : Z.push(e)), !r.is_fork && wt.size > 0 && !Tt && Mt();
	}
	return t;
}
function Mt() {
	Tt = !1;
	for (let e of wt) {
		e.f & 1024 && k(e, y);
		let t;
		try {
			t = On(e);
		} catch {
			t = !0;
		}
		t && Pn(e);
	}
	wt.clear();
}
function Nt(e, t = 1) {
	var n = Q(e), r = t === 1 ? n++ : n--;
	return Ot(e, n), r;
}
function Pt(e) {
	Ot(e, e.v + 1);
}
function Ft(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ve(), a = r.length;
		if (At += a, At > 1e5 && kt === null && (kt = /* @__PURE__ */ new Set()), kt !== null) {
			if (kt.has(e)) return;
			kt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== K) {
				var l = (c & v) === 0;
				if (l && k(s, t), c & 131072) wt.add(s);
				else if (c & 2) {
					var u = s;
					j?.delete(u), Ft(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && M !== null && M.add(d), n === null ? xt(d) : n.push(d);
				}
			}
		}
	}
}
function It(e) {
	if (typeof e != "object" || !e || se in e || ce in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ P(0), s = null, l = Tn, p = (e) => {
		if (Tn === l) return e();
		var t = U, n = Tn;
		G(null), En(l);
		var r = e();
		return G(t), En(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ P(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ne();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ P(n.value, s);
				return i.set(t, e), e;
			}) : Ot(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ P(t, s));
					i.set(n, e), Pt(o);
				}
			} else Ot(r, t), Pt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === se) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ P(It(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = Q(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = Q(a);
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
			if (n === se) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (K !== null || U !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ P(a ? It(e[n]) : t, s)), i.set(n, r)), Q(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ P(t, s)), i.set(f + "", m)) : Ot(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ P(void 0, s)), Ot(u, It(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => It(r));
				Ot(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && Ot(_, v + 1);
				}
				Pt(o);
			}
			return !0;
		},
		ownKeys(e) {
			Q(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Pe();
		}
	});
}
var Lt, Rt, zt, Bt;
function Vt() {
	if (Lt === void 0) {
		Lt = window, Rt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		zt = c(t, "firstChild").get, Bt = c(t, "nextSibling").get, p(e) && (e[de] = void 0, e[ue] = null, e[fe] = void 0, e.__e = void 0), p(n) && (n[pe] = void 0);
	}
}
function F(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ht(e) {
	return zt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function I(e) {
	return Bt.call(e);
}
function L(e, t) {
	if (!w) return /* @__PURE__ */ Ht(e);
	var n = /* @__PURE__ */ Ht(T);
	if (n === null) n = T.appendChild(F());
	else if (t && n.nodeType !== 3) {
		var r = F();
		return n?.before(r), E(r), r;
	}
	return t && qt(n), E(n), n;
}
function Ut(e, t = !1) {
	if (!w) {
		var n = /* @__PURE__ */ Ht(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ I(n) : n;
	}
	if (t) {
		if (T?.nodeType !== 3) {
			var r = F();
			return T?.before(r), E(r), r;
		}
		qt(T);
	}
	return T;
}
function R(e, t = !1) {
	if (!w) return /* @__PURE__ */ Ht(e);
	var n = L(e, t);
	return D(e), n;
}
function z(e, t = 1, n = !1) {
	let r = w ? T : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ I(r);
	if (!w) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = F();
			return r === null ? i?.after(a) : r.before(a), E(a), a;
		}
		qt(r);
	}
	return E(r), r;
}
function Wt(e) {
	e.textContent = "";
}
function Gt() {
	return !1;
}
function Kt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function qt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Jt(e) {
	var t = K;
	if (t === null) return U.f |= oe, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	B(e, t);
}
function B(t, n) {
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
function Yt(e) {
	K === null && (U === null && je(e), Ae()), yn && ke(e);
}
function Xt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Zt(e, t) {
	var n = K;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: O,
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
	A?.register_created_effect(r);
	var i = r;
	if (e & 4) mt === null ? vt.ensure().schedule(r) : mt.push(r);
	else if (t !== null) {
		try {
			Pn(r);
		} catch (e) {
			throw H(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && Xt(i, n), U !== null && U.f & 2 && !(e & 64))) {
		var a = U;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Qt() {
	return U !== null && !W;
}
function $t(e) {
	let t = Zt(8, null);
	return k(t, _), t.teardown = e, t;
}
function en(e) {
	Yt("$effect");
	var t = K.f;
	if (!U && t & 32 && O !== null && !O.i) {
		var n = O;
		(n.e ??= []).push(e);
	} else return tn(e);
}
function tn(e) {
	return Zt(4 | ne, e);
}
function nn(e) {
	vt.ensure();
	let t = Zt(64 | te, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? pn(t, () => {
			H(t), n(void 0);
		}) : (H(t), n(void 0));
	});
}
function rn(e) {
	return Zt(ae | te, e);
}
function an(e, t = 0) {
	return Zt(8 | t, e);
}
function on(e, t = [], n = [], r = []) {
	Ye(r, t, n, (t) => {
		Zt(8, () => {
			e(...t.map(Q));
		});
	});
}
function sn(e, t = 0) {
	return Zt(16 | t, e);
}
function V(e) {
	return Zt(32 | te, e);
}
function cn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = yn, r = U;
		bn(!0), G(null);
		try {
			t.call(null);
		} catch (t) {
			B(t, e.parent);
		} finally {
			bn(n), G(r);
		}
	}
}
function ln(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Je(() => {
			e.abort(me);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : H(n, t), n = r;
	}
}
function un(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || H(t), t = n;
	}
}
function H(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (dn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, ln(e, t && !n), Nn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	cn(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && fn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function dn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ I(e);
		e.remove(), e = n;
	}
}
function fn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function pn(e, t, n = !0) {
	var r = [];
	e.f |= 256, mn(e, r, !0);
	var i = () => {
		n && H(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function mn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				mn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function hn(e) {
	e.f &= -257, gn(e, !0);
}
function gn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (k(e, v), vt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			gn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function _n(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ I(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var vn = null, yn = !1;
function bn(e) {
	yn = e;
}
var U = null, W = !1;
function G(e) {
	U = e;
}
var K = null;
function q(e) {
	K = e;
}
var J = null;
function xn(e) {
	U !== null && (U.f & 2097152 || U.f & 2) && (J ??= /* @__PURE__ */ new Set()).add(e);
}
var Y = null, X = 0, Z = null;
function Sn(e) {
	Z = e;
}
var Cn = 1, wn = 0, Tn = wn;
function En(e) {
	Tn = e;
}
function Dn() {
	return ++Cn;
}
function On(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (On(a) && ot(a), a.wv > e.wv) return !0;
		}
		t & 512 && j === null && k(e, _);
	}
	return !1;
}
function kn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(J !== null && J.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? kn(a, t, !1) : t === a && (n ? k(a, v) : a.f & 1024 && k(a, y), xt(a));
	}
}
function An(e) {
	var t = Y, n = X, r = Z, i = U, a = J, o = O, s = W, c = Tn, l = e.f;
	Y = null, X = 0, Z = null, U = l & 96 ? null : e, J = null, Le(e.ctx), W = !1, Tn = ++wn, e.ac !== null && (Je(() => {
		e.ac.abort(me);
	}), e.ac = null);
	try {
		e.f |= ie;
		var u = e.fn, d = u();
		e.f |= S;
		var f = jn(e);
		if (Ve() && Z !== null && !W && f !== null && !(e.f & 6146)) for (var p = 0; p < Z.length; p++) kn(Z[p], e);
		if (i !== null && i !== e) {
			if (wn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = wn;
			if (t !== null) for (let e of t) e.rv = wn;
			Z !== null && (r === null ? r = Z : r.push(...Z));
		}
		return e.f & 8388608 && (e.f ^= oe), d;
	} catch (t) {
		return jn(e), Jt(t);
	} finally {
		e.f ^= ie, Y = t, X = n, Z = r, U = i, J = a, Le(o), W = s, Tn = c;
	}
}
function jn(e) {
	var t = e.deps, n = A?.is_fork;
	if (Y !== null) {
		var r;
		if (n || Nn(e, X), t !== null && X > 0) for (t.length = X + Y.length, r = 0; r < Y.length; r++) t[X + r] = Y[r];
		else e.deps = t = Y;
		if (Qt() && e.f & 512) for (r = X; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && X < t.length && (Nn(e, X), t.length = X);
	return t;
}
function Mn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (Y === null || !a.call(Y, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ke(c), c.ac !== null && Je(() => {
			c.ac.abort(me), c.ac = null, k(c, v);
		}), st(c), Nn(c, 0);
	}
}
function Nn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Mn(e, n[r]);
}
function Pn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		k(e, _);
		var n = K;
		K = e;
		try {
			t & 16777232 ? un(e) : ln(e), cn(e);
			var r = An(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Cn;
		} finally {
			K = n;
		}
	}
}
function Q(e) {
	var t = !!(e.f & 2);
	if (vn?.add(e), U !== null && !W && !(K !== null && K.f & 16384) && (J === null || !J.has(e))) {
		var n = U.deps;
		if (U.f & 2097152) e.rv < wn && (e.rv = wn, Y === null && n !== null && n[X] === e ? X++ : Y === null ? Y = [e] : Y.push(e));
		else {
			U.deps ??= [], a.call(U.deps, e) || U.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [U] : a.call(r, U) || r.push(U);
		}
	}
	if (yn && N.has(e)) return N.get(e);
	if (t) {
		var i = e;
		if (yn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || In(i)) && (o = at(i)), N.set(i, o), o;
		}
		var s = !(i.f & 512) && !W && U !== null && !!(U.f & 512), c = (i.f & S) === 0;
		On(i) && (s && (i.f |= 512), ot(i)), s && !c && (ct(i), Fn(i));
	}
	if (j?.has(e)) return j.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Fn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (ct(n), Fn(n));
	}
}
function In(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (N.has(t) || t.f & 2 && In(t)) return !0;
	return !1;
}
function Ln(e) {
	var t = W;
	try {
		return W = !0, e();
	} finally {
		W = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Rn = Symbol("events"), zn = /* @__PURE__ */ new Set(), Bn = /* @__PURE__ */ new Set();
function Vn(e, t, n) {
	(t[Rn] ??= {})[e] = n;
}
function Hn(e) {
	for (var t = 0; t < e.length; t++) zn.add(e[t]);
	for (var n of Bn) n(e);
}
var Un = null, Wn = !1;
function Gn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Un = e, Wn || (Wn = !0, setTimeout(() => {
		Wn = !1, Un = null;
	}));
	var o = 0, c = Un === e && e[Rn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Rn] = t;
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
		var d = U, f = K;
		G(null), q(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Rn]?.[r];
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
			e[Rn] = t, delete e.currentTarget, G(d), q(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Kn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function qn(e) {
	return Kn?.createHTML(e) ?? e;
}
function Jn(e) {
	var t = Kt("template");
	return t.innerHTML = qn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Yn(e, t) {
	var n = K;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Xn(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (w) return Yn(T, null), T;
		i === void 0 && (i = Jn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Ht(i)));
		var t = r || Rt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Ht(t), s = t.lastChild;
			Yn(o, s);
		} else Yn(t, t);
		return t;
	};
}
function Zn(e, t) {
	if (w) {
		var n = K;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = T), be();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var Qn = ["touchstart", "touchmove"];
function $n(e) {
	return Qn.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function er(e) {
	let t = 0, n = Et(0), r;
	return () => {
		Qt() && (Q(n), an(() => (t === 0 && (r = Ln(() => e(() => Pt(n)))), t += 1, () => {
			We(() => {
				--t, t === 0 && (r?.(), r = void 0, Pt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var tr = C | te;
function nr(e, t, n, r) {
	new rr(e, t, n, r);
}
var rr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = w ? T : null;
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
	#m = er(() => (this.#p = Et(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = K;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = K.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = sn(() => {
			if (w) {
				let e = Ce(this.#t);
				be();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, tr), w && (this.#e = T);
	}
	#h() {
		try {
			this.#a = V(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		We(r), t && (this.#s = V(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ve() : (t = !0, n && Ie(), this.#s !== null && pn(this.#s, () => {
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
					B(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = V(() => e(this.#e)), We(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = F(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return V(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							B(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(A)) : this.#u === 0 && (this.#e.before(e), this.#c = null, pn(this.#o, () => {
					this.#o = null;
				}), this.#x(A));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = V(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				_n(this.#a, e);
				let t = this.#n.pending;
				this.#o = V(() => t(this.#e));
			} else this.#x(A);
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
		var t = K, n = U, r = O;
		q(this.#i), G(this.#i), Le(this.#i.ctx);
		try {
			return vt.ensure(), e();
		} finally {
			q(t), G(n), Le(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && pn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, We(() => {
			this.#d = !1, this.#p && jt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Q(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		A?.is_fork ? (this.#a && A.skip_effect(this.#a), this.#o && A.skip_effect(this.#o), this.#s && A.skip_effect(this.#s), A.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (H(this.#a), null), this.#o &&= (H(this.#o), null), this.#s &&= (H(this.#s), null), w && (E(this.#t), xe(), E(Se()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return V(() => {
						var r = K;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return B(e, this.#i.parent), null;
				}
			}));
		};
		We(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					B(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => B(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function $(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[pe] ??= e.nodeValue) && (e[pe] = n, e.nodeValue = `${n}`);
}
function ir(e, t) {
	return or(e, t);
}
var ar = /* @__PURE__ */ new Map();
function or(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Vt();
	var u = void 0, d = nn(() => {
		var c = r ?? n.appendChild(F());
		nr(c, { pending: () => {} }, (n) => {
			Re({});
			var r = O;
			if (s && (r.c = s), a && (i.$$events = a), w && Yn(n, null), u = t(n, i) || Be(), w && (K.nodes.end = T, T === null || T.nodeType !== 8 || T.data !== "]")) throw _e(), e;
			ze();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = $n(r);
					for (let e of [n, document]) {
						var a = ar.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), ar.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Gn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(zn)), Bn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = ar.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Gn), t.delete(e), t.size === 0 && ar.delete(r)) : t.set(e, i);
			}
			Bn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return sr.set(u, d), u;
}
var sr = /* @__PURE__ */ new WeakMap();
function cr(e, t) {
	let n = sr.get(e);
	return n ? (sr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var lr = class {
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
			if (n) hn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (hn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (H(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						_n(r, t), t.append(F()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else H(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), pn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (H(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = A, r = Gt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = F();
				i.append(a), this.#n.set(e, {
					effect: V(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, V(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else w && (this.anchor = T), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function ur(e, t, n = !1) {
	var r;
	w && (r = T, be());
	var i = new lr(e), a = n ? C : 0;
	function o(e, t) {
		if (w) {
			var n = Ce(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Se();
				E(a), i.anchor = a, ye(!1), i.ensure(e, t), ye(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	sn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function dr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		pn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					fr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Wt(d), d.append(u), e.items.clear();
		}
		fr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function fr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= re, _n(a, document.createDocumentFragment())) : H(t[i], n);
	}
}
var pr;
function mr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = w ? E(/* @__PURE__ */ Ht(u)) : u.appendChild(F());
	}
	w && be();
	var d = null, f = /* @__PURE__ */ rt(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Q(f);
			_.fallback = d, gr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= re, vr(d, null, c)) : hn(d) : pn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: sn(() => {
			var e = Q(f), r = e.length;
			let o = !1;
			w && Ce(c) === "[!" != (r === 0) && (c = Se(), E(c), ye(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = A, v = Gt(), y = 0; y < r; y += 1) {
				w && T.nodeType === 8 && T.data === "]" && (c = T, o = !0, ye(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && jt(S.v, b), S.i && jt(S.i, y), v && _.unskip_effect(S.e)) : (S = _r(l, m ? c : pr ??= F(), b, x, y, a, t, n), m || (S.e.f |= re), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = V(() => s(c)) : (d = V(() => s(pr ??= F())), d.f |= re)), r > u.size && Oe("", "", ""), w && r > 0 && E(Se()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && ye(!0), Q(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, w && (c = T);
}
function hr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function gr(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = hr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (hn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= re, _ === l) vr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), yr(e, d, _), yr(e, _, y), vr(_, y, n), d = _, p = [], m = [], l = hr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) vr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					yr(e, S.prev, ee.next), yr(e, d, S), yr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), vr(_, l, n), yr(e, _.prev, _.next), yr(e, _, d === null ? e.effect.first : d.next), yr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = hr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = hr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (fr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = hr(l.next);
		var te = C.length;
		if (te > 0) {
			var ne = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.fix();
			}
			dr(e, C, ne);
		}
	}
	a && We(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function _r(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Et(n) : /* @__PURE__ */ Dt(n, !1, !1) : null, l = o & 2 ? Et(i) : null;
	return {
		v: c,
		i: l,
		e: V(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function vr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ I(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function yr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var br = [..." 	\n\r\f\xA0\v﻿"];
function xr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || br.includes(r[o - 1])) && (s === r.length || br.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Sr(e, t, n, r, i, a) {
	var o = e[de];
	if (w || o !== n || o === void 0) {
		var s = xr(n, r, a);
		w && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[de] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Cr = Symbol("is custom element"), wr = Symbol("is html"), Tr = he ? "link" : "LINK";
function Er(e, t, n, r) {
	var i = Dr(e);
	if (!(w && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Tr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[le] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && kr(e).has(t) ? e[t] = n : (!w || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Dr(e) {
	return e[ue] ??= {
		[Cr]: e.nodeName.includes("-"),
		[wr]: e.namespaceURI === n
	};
}
var Or = /* @__PURE__ */ new Map();
function kr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Or.get(t);
	if (n) return n;
	Or.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// kit/plugin.svelte.js
function Ar(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ P(It({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return Q(this.#e);
		}
		set props(e) {
			Ot(this.#e, e, !0);
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
			this.style.display = "contents", this.app ||= ir(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && cr(this.app), this.app = null;
		}
	});
}
function jr(e, t) {
	let n = It({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Nt(r);
		}
	}), r = /* @__PURE__ */ P(0);
	en(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Nt(r);
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
	return en(() => {
		let n = e(), c = t();
		Q(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function Mr(e) {
	let t = It({});
	return en(() => {
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
function Nr(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
function Pr(e) {
	return !e || !e.fields ? {} : typeof e.fields == "string" ? JSON.parse(e.fields) : e.fields;
}
function Fr(e) {
	return !e || !e.meta ? {} : typeof e.meta == "string" ? JSON.parse(e.meta) : e.meta;
}
function Ir(e) {
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
function Lr(e) {
	let t = Fr(e), n = [], r = t.github || {}, i = r.repo || "elee1766/jobshop";
	r.issue && n.push({
		label: "issue #" + r.issue,
		href: `https://github.com/${i}/issues/${r.issue}`
	}), r.pr && n.push({
		label: "PR #" + r.pr,
		href: `https://github.com/${i}/pull/${r.pr}`
	});
	let a = t.linear || {}, o = a.issue || a.identifier, s = a.url || (a.workspace && o ? `https://linear.app/${encodeURIComponent(a.workspace)}/issue/${encodeURIComponent(o)}` : "");
	return s && Rr(s) && n.push({
		label: o || "Linear",
		href: s
	}), n;
}
function Rr(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// plugins/work/Work.svelte
var zr = /* @__PURE__ */ Xn("<button> </button>"), Br = /* @__PURE__ */ Xn("<div class=\"card err\"> </div>"), Vr = /* @__PURE__ */ Xn("<tr class=\"click\"><td class=\"mono c-id\"> </td><td class=\"c-title\"> </td><td class=\"muted c-time\"> </td></tr>"), Hr = /* @__PURE__ */ Xn("<h2 class=\"st\">Batches</h2> <div class=\"card scroll\"><table class=\"stack\"><tbody></tbody></table></div>", 1), Ur = /* @__PURE__ */ Xn("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"svelte-1k64nxg\"> </a>"), Wr = /* @__PURE__ */ Xn("<tr class=\"click\"><td class=\"c-title\"> <div class=\"faint mono\"> </div></td><td class=\"c-id\"><span> </span></td><td class=\"muted\"> </td><td class=\"links svelte-1k64nxg\"></td><td class=\"muted c-time\"> </td></tr>"), Gr = /* @__PURE__ */ Xn("<tr><td colspan=\"5\" class=\"faint\">No runs.</td></tr>"), Kr = /* @__PURE__ */ Xn("<!> <h2 class=\"st\">Runs</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"runs\"><thead><tr><th>Run</th><th>State</th><th>Workflow</th><th>Links</th><th>Updated</th></tr></thead><tbody></tbody></table></div>", 1), qr = /* @__PURE__ */ Xn("<header class=\"vh svelte-1k64nxg\"><h1>Work</h1> <div class=\"chips actions svelte-1k64nxg\" role=\"group\" aria-label=\"state filter\"></div></header> <div class=\"pad\"><!></div>", 1);
function Jr(e, t) {
	Re(t, !0);
	let n = Mr(() => t.api), r = (e) => t.api.go(e), i = [
		"",
		"running",
		"needs_human",
		"done",
		"cancelled"
	], a = jr(() => t.api, () => "api/work" + (n.state ? "?state=" + encodeURIComponent(n.state) : ""));
	var o = qr(), s = Ut(o), c = z(L(s), 2);
	mr(c, 20, () => i, (e) => e, (e, t) => {
		var i = zr();
		let a;
		var o = R(i, !0);
		on((e) => {
			a = Sr(i, 1, "", null, a, { on: n.state === t }), $(o, e);
		}, [() => t ? t.replace("_", " ") : "all"]), Vn("click", i, () => r({ state: t })), Zn(e, i);
	}), D(c), D(s);
	var l = z(s, 2), u = L(l), d = (e) => {
		var t = Br(), n = R(t, !0);
		on(() => $(n, a.error)), Zn(e, t);
	}, f = (e) => {
		var t = Kr(), n = Ut(t), i = (e) => {
			var t = Hr(), n = z(Ut(t), 2), i = L(n), o = L(i);
			mr(o, 21, () => a.data.batches, (e) => e.id, (e, t) => {
				var n = Vr(), i = L(n), a = R(i, !0), o = z(i), s = R(o, !0), c = R(z(o), !0);
				D(n), on((e) => {
					$(a, Q(t).id), $(s, Q(t).title), $(c, e);
				}, [() => Nr(Q(t).updated_at)]), Vn("click", n, () => r({ ticket: Q(t).id })), Zn(e, n);
			}), D(o), D(i), D(n), Zn(e, t);
		};
		ur(n, (e) => {
			a.data.batches.length && e(i);
		});
		var o = z(n, 4), s = L(o), c = z(L(s));
		mr(c, 21, () => a.data.runs, (e) => e.id, (e, t) => {
			let n = /* @__PURE__ */ nt(() => Pr(Q(t)));
			var i = Wr(), a = L(i), o = L(a, !0), s = R(z(o));
			D(a);
			var c = z(a), l = L(c), u = R(l, !0);
			D(c);
			var d = z(c), f = R(d, !0), p = z(d);
			mr(p, 21, () => Lr(Q(t)), (e) => e.href, (e, t) => {
				var n = Ur(), r = R(n, !0);
				on(() => {
					Er(n, "href", Q(t).href), $(r, Q(t).label);
				}), Vn("click", n, (e) => e.stopPropagation()), Zn(e, n);
			}), D(p);
			var m = R(z(p), !0);
			D(i), on((e, r) => {
				$(o, Q(t).title), $(s, `${Q(t).id ?? ""} · ${Q(t).project ?? ""}`), Sr(l, 1, `pill ${e ?? ""}`), $(u, Q(n).state || Q(t).status), $(f, Q(n).workflow || ""), $(m, r);
			}, [() => Ir(Q(n).state), () => Nr(Q(t).updated_at)]), Vn("click", i, () => r({ ticket: Q(t).id })), Zn(e, i);
		}, (e) => {
			Zn(e, Gr());
		}), D(c), D(s), D(o), Zn(e, t);
	};
	ur(u, (e) => {
		a.error ? e(d) : a.data && e(f, 1);
	}), D(l), Zn(e, o), ze();
}
// plugins/work/index.js
Hn(["click"]), Ar("desk-work-view", Jr);
