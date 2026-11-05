// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = Array.isArray, r = Array.prototype.indexOf, i = Array.prototype.includes, a = Array.from, o = Object.defineProperty, s = Object.getOwnPropertyDescriptor, c = Object.prototype, l = Array.prototype, u = Object.getPrototypeOf, d = Object.isExtensible, f = () => {};
function p(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function m() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
function h(e, t) {
	if (Array.isArray(e)) return e;
	if (t === void 0 || !(Symbol.iterator in e)) return Array.from(e);
	let n = [];
	for (let r of e) if (n.push(r), n.length === t) break;
	return n;
}
var g = 1024, _ = 2048, v = 4096, y = 8192, b = 16384, x = 32768, S = 1 << 25, C = 65536, w = 1 << 19, ee = 1 << 20, te = 1 << 25, T = 1 << 21, ne = 1 << 22, re = 1 << 23, ie = Symbol("$state"), ae = Symbol("component"), oe = Symbol("attributes"), se = Symbol("class"), ce = Symbol("style"), le = Symbol("text"), ue = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}();
globalThis.document?.contentType;
function de() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function fe(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function pe() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var E = !1;
function me(e) {
	E = e;
}
var D;
function he(t) {
	if (t === null) throw fe(), e;
	return D = t;
}
function ge() {
	return he(/* @__PURE__ */ Bt(D));
}
function O(t) {
	if (E) {
		if (/* @__PURE__ */ Bt(D) !== null) throw fe(), e;
		D = t;
	}
}
function _e(e = 1) {
	if (E) {
		for (var t = e, n = D; t--;) n = /* @__PURE__ */ Bt(n);
		D = n;
	}
}
function ve(e = !0) {
	for (var t = 0, n = D;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Bt(n);
		e && n.remove(), n = i;
	}
}
function ye(t) {
	if (!t || t.nodeType !== 8) throw fe(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function be(e) {
	return e === this.v;
}
function xe(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Se(e) {
	return !xe(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function k() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Ce(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function A(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function we() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Te(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Ee() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function De() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Oe() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function ke() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Ae() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var j = null;
function je(e) {
	j = e;
}
function Me(e, t = !1, n) {
	j = {
		p: j,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: V,
		l: null
	};
}
function Ne(e) {
	var t = j, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) $t(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, j = t.p, Pe(e);
}
function Pe(e = {}) {
	return o(e, ae, { value: !0 }), e;
}
function Fe() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var Ie = [];
function Le() {
	var e = Ie;
	Ie = [], p(e);
}
function Re(e) {
	if (Ie.length === 0 && !st) {
		var t = Ie;
		queueMicrotask(() => {
			t === Ie && Le();
		});
	}
	Ie.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var ze = ~(_ | v | g);
function M(e, t) {
	e.f = e.f & ze | t;
}
function Be(e) {
	e.f & 512 || e.deps === null ? M(e, g) : M(e, v);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ve(e, t) {
	e.f & 2048 ? t.set(e, _) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, v), M(e, g);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function He(e) {
	var t = B, n = V;
	Sn(null), Cn(null);
	try {
		return e();
	} finally {
		Sn(t), Cn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ue(e, t, n, r) {
	let i = Fe() ? qe : Ze;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = V, c = We(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				z(e, s);
			}
			Ge();
		}
	}
	var d = Ke();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ Ye(e))).then(u).catch((e) => z(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Ge());
	}) : f();
}
function We() {
	var e = V, t = B, n = j, r = N;
	return function(i = !0) {
		Cn(e), Sn(t), je(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Ge(e = !0) {
	Cn(null), Sn(null), je(null), e && N?.deactivate();
}
function Ke() {
	var e = V, t = e.b, n = N, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function qe(e) {
	var n = 2 | _;
	return V !== null && (V.f |= w), {
		ctx: j,
		deps: null,
		effects: null,
		equals: be,
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
var Je = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function Ye(e, n, r) {
	let i = V;
	i === null && k();
	var a = void 0, o = St(t), s = !B, c = /* @__PURE__ */ new Set();
	return tn(() => {
		var t = V, n = m();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ue && n.reject(e);
			}).finally(Ge);
		} catch (e) {
			n.reject(e), Ge();
		}
		var r = N;
		if (s) {
			if (t.f & 32768) var l = Ke();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(Je);
			else for (let e of c.values()) e.reject(Je);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== Je && (r.activate(), t ? (o.f |= re, Dt(o, t)) : (o.f & 8388608 && (o.f ^= re), Dt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), Zt(() => {
		for (let e of c) e.reject(Je);
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
function Xe(e) {
	let t = /* @__PURE__ */ qe(e);
	return Tn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function Ze(e) {
	let t = /* @__PURE__ */ qe(e);
	return t.equals = Se, t;
}
function Qe(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) un(t[n]);
	}
}
function $e(e) {
	var n, r = V, i = e.parent;
	if (!yn && i !== null && e.v !== t && i.f & 24576) return de(), e.v;
	Cn(i);
	try {
		Qe(e), n = Ln(e);
	} finally {
		Cn(r);
	}
	return n;
}
function et(e) {
	var t = $e(e);
	!e.equals(t) && (e.wv = Pn(), (!N?.is_fork || e.deps === null) && (N === null ? e.v = t : (N.capture(e, t, !0), it?.capture(e, t, !0)), e.deps === null)) ? M(e, g) : yn || (at === null ? Be(e) : (Xt() || N?.is_fork) && at.set(e, t));
}
function tt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && He(() => {
		t.ac.abort(ue), t.ac = null;
	}), t.fn !== null && (t.teardown = f), Bn(t, 0), cn(t));
}
function nt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Vn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var rt = null, N = null, it = null, at = null, ot = null, st = !1, ct = !1, lt = null, ut = null, dt = 0, ft = 1, P = class e {
	id = ft++;
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
		rt === null ? rt = this : (rt.next = this, this.prev = rt), rt = this;
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
			for (var r of n.d) M(r, _), t(r);
			for (r of n.m) M(r, v), t(r);
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
					t.f ^= g;
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
		for (var t = lt = [], n = [], r = ut = []; this.#a.length > 0;) {
			dt++ > 1e3 && (this.#b(), pt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw vt(e), this.#d() || this.discard(), t;
			}
		}
		if (N = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (lt = null, ut = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) _t(e, t);
			r.length > 0 && N.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), it = this, ht(n), ht(t), it = null, this.#i?.resolve();
			var o = N;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (bt.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= g;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= g : i & 4 ? t.push(r) : Fn(r) && (i & 16 && this.add_dirty_reaction(r, v), Vn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, _), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), N = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Ve(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), at?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		N = this;
	}
	deactivate() {
		N = null, at = null;
	}
	flush() {
		try {
			ct = !0, N = this, this.#m();
		} finally {
			dt = 0, ot = null, lt = null, ut = null, ct = !1, N = null, at = null, bt.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(Je);
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
		this.#u || (this.#u = !0, Re(() => {
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
		return (this.#i ??= m()).promise;
	}
	static ensure() {
		if (N === null) {
			let t = N = new e();
			!ct && Re(() => {
				t.started || t.flush();
			});
		}
		return N;
	}
	apply() {
		at = null;
	}
	schedule(e) {
		ot = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? rt = e : t.prev = e, this.linked = !1;
		}
	}
};
function pt() {
	try {
		Ee();
	} catch (e) {
		z(e, ot);
	}
}
var mt = null;
function ht(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Fn(r) && (mt = /* @__PURE__ */ new Set(), Vn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && fn(r), mt?.size > 0)) {
				bt.clear();
				for (let e of mt) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) mt.has(n) && (mt.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Vn(n);
					}
				}
				mt.clear();
			}
		}
		mt = null;
	}
}
function gt(e) {
	N.schedule(e);
}
function _t(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, g);
		for (var n = e.first; n !== null;) _t(n, t), n = n.next;
	}
}
function vt(e) {
	M(e, g);
	for (var t = e.first; t !== null;) vt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var yt = /* @__PURE__ */ new Set(), bt = /* @__PURE__ */ new Map(), xt = !1;
function St(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: be,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function F(e, t) {
	let n = St(e, t);
	return Tn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Ct(e, t = !1, n = !0) {
	let r = St(e);
	return t || (r.equals = Se), r;
}
function wt(e, t, n = !1) {
	return B !== null && (!xn || B.f & 131072) && Fe() && B.f & 4325394 && (wn === null || !wn.has(e)) && ke(), Dt(e, n ? Mt(t) : t, ut);
}
var Tt = null, Et = 0;
function Dt(e, t, n = null) {
	if (!e.equals(t)) {
		yn ? bt.set(e, t) : bt.has(e) || bt.set(e, e.v);
		var r = P.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && $e(t), at === null && Be(t);
		}
		e.wv = Pn(), Tt = null, Et = 0, jt(e, _, n), Tt = null, Fe() && V !== null && V.f & 1024 && !(V.f & 96) && (On === null ? kn([e]) : On.push(e)), !r.is_fork && yt.size > 0 && !xt && Ot();
	}
	return t;
}
function Ot() {
	xt = !1;
	for (let e of yt) {
		e.f & 1024 && M(e, v);
		let t;
		try {
			t = Fn(e);
		} catch {
			t = !0;
		}
		t && Vn(e);
	}
	yt.clear();
}
function kt(e, t = 1) {
	var n = H(e), r = t === 1 ? n++ : n--;
	return wt(e, n), r;
}
function At(e) {
	wt(e, e.v + 1);
}
function jt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Fe(), a = r.length;
		if (Et += a, Et > 1e5 && Tt === null && (Tt = /* @__PURE__ */ new Set()), Tt !== null) {
			if (Tt.has(e)) return;
			Tt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== V) {
				var l = (c & _) === 0;
				if (l && M(s, t), c & 131072) yt.add(s);
				else if (c & 2) {
					var u = s;
					at?.delete(u), jt(u, v, n);
				} else if (l) {
					var d = s;
					c & 16 && mt !== null && mt.add(d), n === null ? gt(d) : n.push(d);
				}
			}
		}
	}
}
function Mt(e) {
	if (typeof e != "object" || !e || ie in e || ae in e) return e;
	let r = u(e);
	if (r !== c && r !== l) return e;
	var i = /* @__PURE__ */ new Map(), a = n(e), o = /* @__PURE__ */ F(0), d = null, f = Mn, p = (e) => {
		if (Mn === f) return e();
		var t = B, n = Mn;
		Sn(null), Nn(f);
		var r = e();
		return Sn(t), Nn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ F(e.length, d)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || De();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ F(n.value, d);
				return i.set(t, e), e;
			}) : wt(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ F(t, d));
					i.set(n, e), At(o);
				}
			} else wt(r, t), At(o);
			return !0;
		},
		get(n, r, a) {
			if (r === ie) return e;
			var o = i.get(r), c = r in n;
			if (o === void 0 && (!c || s(n, r)?.writable) && (o = p(() => /* @__PURE__ */ F(Mt(c ? n[r] : t), d)), i.set(r, o)), o !== void 0) {
				var l = H(o);
				return l === t ? void 0 : l;
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
			if (n === ie) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (V !== null || B !== null) && (!a || s(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ F(a ? Mt(e[n]) : t, d)), i.set(n, r)), H(r) === t) ? !1 : a;
		},
		set(e, n, r, c) {
			var l = i.get(n), u = n in e;
			if (a && n === "length") for (var f = r; f < l.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ F(t, d)), i.set(f + "", m)) : wt(m, t);
			}
			if (l === void 0) (!u || s(e, n)?.writable) && (l = p(() => /* @__PURE__ */ F(void 0, d)), wt(l, Mt(r)), i.set(n, l));
			else {
				u = l.v !== t;
				var h = p(() => Mt(r));
				wt(l, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, r), !u) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && wt(_, v + 1);
				}
				At(o);
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
			Oe();
		}
	});
}
var Nt, Pt, Ft, It;
function Lt() {
	if (Nt === void 0) {
		Nt = window, Pt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Ft = s(t, "firstChild").get, It = s(t, "nextSibling").get, d(e) && (e[se] = void 0, e[oe] = null, e[ce] = void 0, e.__e = void 0), d(n) && (n[le] = void 0);
	}
}
function Rt(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function zt(e) {
	return Ft.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Bt(e) {
	return It.call(e);
}
function I(e, t) {
	if (!E) return /* @__PURE__ */ zt(e);
	var n = /* @__PURE__ */ zt(D);
	if (n === null) n = D.appendChild(Rt());
	else if (t && n.nodeType !== 3) {
		var r = Rt();
		return n?.before(r), he(r), r;
	}
	return t && Gt(n), he(n), n;
}
function Vt(e, t = !1) {
	if (!E) {
		var n = /* @__PURE__ */ zt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Bt(n) : n;
	}
	if (t) {
		if (D?.nodeType !== 3) {
			var r = Rt();
			return D?.before(r), he(r), r;
		}
		Gt(D);
	}
	return D;
}
function L(e, t = !1) {
	if (!E) return /* @__PURE__ */ zt(e);
	var n = I(e, t);
	return O(e), n;
}
function R(e, t = 1, n = !1) {
	let r = E ? D : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Bt(r);
	if (!E) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = Rt();
			return r === null ? i?.after(a) : r.before(a), he(a), a;
		}
		Gt(r);
	}
	return he(r), r;
}
function Ht(e) {
	e.textContent = "";
}
function Ut() {
	return !1;
}
function Wt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Gt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function Kt(e) {
	var t = V;
	if (t === null) return B.f |= re, e;
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
function qt(e) {
	V === null && (B === null && Te(e), we()), yn && A(e);
}
function Jt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Yt(e, t) {
	var n = V;
	n !== null && n.f & 8192 && (e |= y);
	var r = {
		ctx: j,
		deps: null,
		nodes: null,
		f: e | _ | 512,
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
	if (e & 4) lt === null ? P.ensure().schedule(r) : lt.push(r);
	else if (t !== null) {
		try {
			Vn(r);
		} catch (e) {
			throw un(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && Jt(i, n), B !== null && B.f & 2 && !(e & 64))) {
		var a = B;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Xt() {
	return B !== null && !xn;
}
function Zt(e) {
	let t = Yt(8, null);
	return M(t, g), t.teardown = e, t;
}
function Qt(e) {
	qt("$effect");
	var t = V.f;
	if (!B && t & 32 && j !== null && !j.i) {
		var n = j;
		(n.e ??= []).push(e);
	} else return $t(e);
}
function $t(e) {
	return Yt(4 | ee, e);
}
function en(e) {
	P.ensure();
	let t = Yt(64 | w, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? pn(t, () => {
			un(t), n(void 0);
		}) : (un(t), n(void 0));
	});
}
function tn(e) {
	return Yt(ne | w, e);
}
function nn(e, t = 0) {
	return Yt(8 | t, e);
}
function rn(e, t = [], n = [], r = []) {
	Ue(r, t, n, (t) => {
		Yt(8, () => {
			e(...t.map(H));
		});
	});
}
function an(e, t = 0) {
	return Yt(16 | t, e);
}
function on(e) {
	return Yt(32 | w, e);
}
function sn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = yn, r = B;
		bn(!0), Sn(null);
		try {
			t.call(null);
		} catch (t) {
			z(t, e.parent);
		} finally {
			bn(n), Sn(r);
		}
	}
}
function cn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && He(() => {
			e.abort(ue);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : un(n, t), n = r;
	}
}
function ln(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || un(t), t = n;
	}
}
function un(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (dn(e.nodes.start, e.nodes.end), n = !0), e.f |= S, cn(e, t && !n), Bn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	sn(e), e.f ^= S, e.f |= b;
	var i = e.parent;
	i !== null && i.first !== null && fn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function dn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Bt(e);
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
		n && un(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function mn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= y;
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
		e.f ^= y, e.f & 1024 || (M(e, _), P.ensure().schedule(e));
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
		var i = n === r ? null : /* @__PURE__ */ Bt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var vn = null, yn = !1;
function bn(e) {
	yn = e;
}
var B = null, xn = !1;
function Sn(e) {
	B = e;
}
var V = null;
function Cn(e) {
	V = e;
}
var wn = null;
function Tn(e) {
	B !== null && (B.f & 2097152 || B.f & 2) && (wn ??= /* @__PURE__ */ new Set()).add(e);
}
var En = null, Dn = 0, On = null;
function kn(e) {
	On = e;
}
var An = 1, jn = 0, Mn = jn;
function Nn(e) {
	Mn = e;
}
function Pn() {
	return ++An;
}
function Fn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Fn(a) && et(a), a.wv > e.wv) return !0;
		}
		t & 512 && at === null && M(e, g);
	}
	return !1;
}
function In(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(wn !== null && wn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? In(a, t, !1) : t === a && (n ? M(a, _) : a.f & 1024 && M(a, v), gt(a));
	}
}
function Ln(e) {
	var t = En, n = Dn, r = On, i = B, a = wn, o = j, s = xn, c = Mn, l = e.f;
	En = null, Dn = 0, On = null, B = l & 96 ? null : e, wn = null, je(e.ctx), xn = !1, Mn = ++jn, e.ac !== null && (He(() => {
		e.ac.abort(ue);
	}), e.ac = null);
	try {
		e.f |= T;
		var u = e.fn, d = u();
		e.f |= x;
		var f = Rn(e);
		if (Fe() && On !== null && !xn && f !== null && !(e.f & 6146)) for (var p = 0; p < On.length; p++) In(On[p], e);
		if (i !== null && i !== e) {
			if (jn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = jn;
			if (t !== null) for (let e of t) e.rv = jn;
			On !== null && (r === null ? r = On : r.push(...On));
		}
		return e.f & 8388608 && (e.f ^= re), d;
	} catch (t) {
		return Rn(e), Kt(t);
	} finally {
		e.f ^= T, En = t, Dn = n, On = r, B = i, wn = a, je(o), xn = s, Mn = c;
	}
}
function Rn(e) {
	var t = e.deps, n = N?.is_fork;
	if (En !== null) {
		var r;
		if (n || Bn(e, Dn), t !== null && Dn > 0) for (t.length = Dn + En.length, r = 0; r < En.length; r++) t[Dn + r] = En[r];
		else e.deps = t = En;
		if (Xt() && e.f & 512) for (r = Dn; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Dn < t.length && (Bn(e, Dn), t.length = Dn);
	return t;
}
function zn(e, n) {
	let a = n.reactions;
	if (a !== null) {
		var o = r.call(a, e);
		if (o !== -1) {
			var s = a.length - 1;
			s === 0 ? a = n.reactions = null : (a[o] = a[s], a.pop());
		}
	}
	if (a === null && n.f & 2 && (En === null || !i.call(En, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Be(c), c.ac !== null && He(() => {
			c.ac.abort(ue), c.ac = null, M(c, _);
		}), tt(c), Bn(c, 0);
	}
}
function Bn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) zn(e, n[r]);
}
function Vn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		M(e, g);
		var n = V;
		V = e;
		try {
			t & 16777232 ? ln(e) : cn(e), sn(e);
			var r = Ln(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = An;
		} finally {
			V = n;
		}
	}
}
function H(e) {
	var t = !!(e.f & 2);
	if (vn?.add(e), B !== null && !xn && !(V !== null && V.f & 16384) && (wn === null || !wn.has(e))) {
		var n = B.deps;
		if (B.f & 2097152) e.rv < jn && (e.rv = jn, En === null && n !== null && n[Dn] === e ? Dn++ : En === null ? En = [e] : En.push(e));
		else {
			B.deps ??= [], i.call(B.deps, e) || B.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [B] : i.call(r, B) || r.push(B);
		}
	}
	if (yn && bt.has(e)) return bt.get(e);
	if (t) {
		var a = e;
		if (yn) {
			var o = a.v;
			return (!(a.f & 1024) && a.reactions !== null || Un(a)) && (o = $e(a)), bt.set(a, o), o;
		}
		var s = !(a.f & 512) && !xn && B !== null && !!(B.f & 512), c = (a.f & x) === 0;
		Fn(a) && (s && (a.f |= 512), et(a)), s && !c && (nt(a), Hn(a));
	}
	if (at?.has(e)) return at.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Hn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : i.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (nt(n), Hn(n));
	}
}
function Un(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (bt.has(t) || t.f & 2 && Un(t)) return !0;
	return !1;
}
function Wn(e) {
	var t = xn;
	try {
		return xn = !0, e();
	} finally {
		xn = t;
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
	var s = 0, c = Xn === e && e[Gn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Gn] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (s = l);
	}
	if (a = i[s] || e.target, a !== t) {
		o(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = B, f = V;
		Sn(null), Cn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Gn]?.[r];
					h == null || a.disabled && e.target !== a || h.call(a, e);
				} catch (e) {
					p ? m.push(e) : p = e;
				}
				if (e.cancelBubble) break;
				s++, a = s < i.length ? i[s] : null;
			}
			if (p) {
				for (let e of m) queueMicrotask(() => {
					throw e;
				});
				throw p;
			}
		} finally {
			e[Gn] = t, delete e.currentTarget, Sn(d), Cn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var $n = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function er(e) {
	return $n?.createHTML(e) ?? e;
}
function tr(e) {
	var t = Wt("template");
	return t.innerHTML = er(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function nr(e, t) {
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
		if (E) return nr(D, null), D;
		i === void 0 && (i = tr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ zt(i)));
		var t = r || Pt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ zt(t), s = t.lastChild;
			nr(o, s);
		} else nr(t, t);
		return t;
	};
}
function W(e, t) {
	if (E) {
		var n = V;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = D), ge();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var rr = ["touchstart", "touchmove"];
function ir(e) {
	return rr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function ar(e) {
	let t = 0, n = St(0), r;
	return () => {
		Xt() && (H(n), nn(() => (t === 0 && (r = Wn(() => e(() => At(n)))), t += 1, () => {
			Re(() => {
				--t, t === 0 && (r?.(), r = void 0, At(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var or = C | w;
function sr(e, t, n, r) {
	new cr(e, t, n, r);
}
var cr = class {
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
	#m = ar(() => (this.#p = St(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = V;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = V.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = an(() => {
			if (E) {
				let e = ye(this.#t);
				ge();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, or), E && (this.#e = D);
	}
	#h() {
		try {
			this.#a = on(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Re(r), t && (this.#s = on(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? pe() : (t = !0, n && Ae(), this.#s !== null && pn(this.#s, () => {
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
		return (this.#i.f & (b | S)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = on(() => e(this.#e)), Re(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = Rt(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return on(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							z(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(N)) : this.#u === 0 && (this.#e.before(e), this.#c = null, pn(this.#o, () => {
					this.#o = null;
				}), this.#x(N));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = on(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				_n(this.#a, e);
				let t = this.#n.pending;
				this.#o = on(() => t(this.#e));
			} else this.#x(N);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ve(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = V, n = B, r = j;
		Cn(this.#i), Sn(this.#i), je(this.#i.ctx);
		try {
			return P.ensure(), e();
		} finally {
			Cn(t), Sn(n), je(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && pn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Re(() => {
			this.#d = !1, this.#p && Dt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), H(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		N?.is_fork ? (this.#a && N.skip_effect(this.#a), this.#o && N.skip_effect(this.#o), this.#s && N.skip_effect(this.#s), N.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (un(this.#a), null), this.#o &&= (un(this.#o), null), this.#s &&= (un(this.#s), null), E && (he(this.#t), _e(), he(ve()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return on(() => {
						var r = V;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return z(e, this.#i.parent), null;
				}
			}));
		};
		Re(() => {
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
function G(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[le] ??= e.nodeValue) && (e[le] = n, e.nodeValue = `${n}`);
}
function lr(e, t) {
	return dr(e, t);
}
var ur = /* @__PURE__ */ new Map();
function dr(t, { target: n, anchor: r, props: i = {}, events: o, context: s, intro: c = !0, transformError: l }) {
	Lt();
	var u = void 0, d = en(() => {
		var c = r ?? n.appendChild(Rt());
		sr(c, { pending: () => {} }, (n) => {
			Me({});
			var r = j;
			if (s && (r.c = s), o && (i.$$events = o), E && nr(n, null), u = t(n, i) || Pe(), E && (V.nodes.end = D, D === null || D.nodeType !== 8 || D.data !== "]")) throw fe(), e;
			Ne();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = ir(r);
					for (let e of [n, document]) {
						var a = ur.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), ur.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Qn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(a(Kn)), qn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = ur.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Qn), t.delete(e), t.size === 0 && ur.delete(r)) : t.set(e, i);
			}
			qn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return fr.set(u, d), u;
}
var fr = /* @__PURE__ */ new WeakMap();
function pr(e, t) {
	let n = fr.get(e);
	return n ? (fr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var mr = class {
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
				r && (un(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						_n(r, t), t.append(Rt()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else un(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), pn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (un(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = N, r = Ut();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = Rt();
				i.append(a), this.#n.set(e, {
					effect: on(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, on(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else E && (this.anchor = D), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function hr(e, t, n = !1) {
	var r;
	E && (r = D, ge());
	var i = new mr(e), a = n ? C : 0;
	function o(e, t) {
		if (E) {
			var n = ye(r);
			if (e !== parseInt(n.substring(1))) {
				var a = ve();
				he(a), i.anchor = a, me(!1), i.ensure(e, t), me(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	an(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function gr(e, t, n) {
	for (var r = [], i = t.length, o, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		pn(n, () => {
			if (o) {
				if (o.pending.delete(n), o.done.add(n), o.pending.size === 0) {
					var t = e.outrogroups;
					_r(e, a(o.done)), t.delete(o), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Ht(d), d.append(u), e.items.clear();
		}
		_r(e, t, !l);
	} else o = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(o);
}
function _r(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= te, _n(a, document.createDocumentFragment())) : un(t[i], n);
	}
}
var vr;
function yr(e, t, r, i, o, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = E ? he(/* @__PURE__ */ zt(u)) : u.appendChild(Rt());
	}
	E && ge();
	var d = null, f = /* @__PURE__ */ Ze(() => {
		var e = r();
		return n(e) ? e : e == null ? [] : a(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = H(f);
			_.fallback = d, xr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= te, Cr(d, null, c)) : hn(d) : pn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: an(() => {
			var e = H(f), n = e.length;
			let a = !1;
			E && ye(c) === "[!" != (n === 0) && (c = ve(), he(c), me(!1), a = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = N, v = Ut(), y = 0; y < n; y += 1) {
				E && D.nodeType === 8 && D.data === "]" && (c = D, a = !0, me(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Dt(S.v, b), S.i && Dt(S.i, y), v && _.unskip_effect(S.e)) : (S = Sr(l, m ? c : vr ??= Rt(), b, x, y, o, t, r), m || (S.e.f |= te), l.set(x, S)), u.add(x);
			}
			if (n === 0 && s && !d && (m ? d = on(() => s(c)) : (d = on(() => s(vr ??= Rt())), d.f |= te)), n > u.size && Ce("", "", ""), E && n > 0 && he(ve()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			a && me(!0), H(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, E && (c = D);
}
function br(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function xr(e, t, n, r, i) {
	var o = !!(r & 8), s = t.length, c = e.items, l = br(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (o) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (hn(_), o && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= te, _ === l) Cr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), wr(e, d, _), wr(e, _, y), Cr(_, y, n), d = _, p = [], m = [], l = br(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], C = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Cr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					wr(e, S.prev, C.next), wr(e, d, S), wr(e, C, b), l = b, d = C, --v, p = [], m = [];
				} else u.delete(_), Cr(_, l, n), wr(e, _.prev, _.next), wr(e, _, d === null ? e.effect.first : d.next), wr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = br(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = br(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (_r(e, a(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = br(l.next);
		var ee = w.length;
		if (ee > 0) {
			var T = r & 4 && s === 0 ? n : null;
			if (o) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			gr(e, w, T);
		}
	}
	o && Re(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Sr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? St(n) : /* @__PURE__ */ Ct(n, !1, !1) : null, l = o & 2 ? St(i) : null;
	return {
		v: c,
		i: l,
		e: on(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Cr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Bt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function wr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var Tr = [..." 	\n\r\f\xA0\v﻿"];
function Er(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Tr.includes(r[o - 1])) && (s === r.length || Tr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Dr(e, t, n, r, i, a) {
	var o = e[se];
	if (E || o !== n || o === void 0) {
		var s = Er(n, r, a);
		E && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[se] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// kit/plugin.svelte.js
function Or(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ F(Mt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return H(this.#e);
		}
		set props(e) {
			wt(this.#e, e, !0);
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
			this.style.display = "contents", this.app ||= lr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && pr(this.app), this.app = null;
		}
	});
}
function kr(e, t) {
	let n = Mt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			kt(r);
		}
	}), r = /* @__PURE__ */ F(0);
	Qt(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, kt(r);
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
	return Qt(() => {
		let n = e(), c = t();
		H(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
var Ar = (e) => String(e?.message || e);
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/marked/lib/marked.esm.js
function jr() {
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
var Mr = jr();
function Nr(e) {
	Mr = e;
}
var Pr = { exec: () => null };
function Fr(e) {
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
var Ir = ((e = "") => {
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
	nextBulletRegex: Fr((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Fr((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Fr((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Fr((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Fr((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Fr((e) => RegExp(`^ {0,${e}}(?:</?(?:${Zr})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Fr((e) => RegExp(`^ {0,${e}}>`))
}, Lr = /^(?:[ \t]*(?:\n|$))+/, Rr = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, zr = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, Br = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, Vr = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, Hr = / {0,3}(?:[*+-]|\d{1,9}[.)])/, Ur = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, Wr = K(Ur).replace(/bull/g, Hr).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Gr = K(Ur).replace(/bull/g, Hr).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), Kr = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, qr = /^[^\n]+/, Jr = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Yr = K(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", Jr).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), Xr = K(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, Hr).getRegex(), Zr = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Qr = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, $r = K("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Qr).replace("tag", Zr).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), ei = (e) => K(Kr).replace("hr", Br).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Zr).getRegex(), ti = ei(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), ni = ei(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), ri = {
	blockquote: K(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", ni).getRegex(),
	code: Rr,
	def: Yr,
	fences: zr,
	heading: Vr,
	hr: Br,
	html: $r,
	lheading: Wr,
	list: Xr,
	newline: Lr,
	paragraph: ti,
	table: Pr,
	text: qr
}, ii = K("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", Br).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Zr).getRegex(), ai = {
	...ri,
	lheading: Gr,
	table: ii,
	paragraph: K(Kr).replace("hr", Br).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", ii).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Zr).getRegex()
}, oi = {
	...ri,
	html: K("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Qr).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: Pr,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: K(Kr).replace("hr", Br).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Wr).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, si = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, ci = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, li = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, ui = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, di = /[\p{P}\p{S}]/u, fi = /[\s\p{P}\p{S}]/u, pi = /[^\s\p{P}\p{S}]/u, mi = K(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, fi).getRegex(), hi = /[\p{Pi}\p{Ps}"']/u, gi = /(?!~)[\p{P}\p{S}]/u, _i = /(?!~)[\s\p{P}\p{S}]/u, vi = /(?:[^\s\p{P}\p{S}]|~)/u, yi = K(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", Ir ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), bi = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, xi = K(bi, "u").replace(/punct/g, di).getRegex(), Si = K(bi, "u").replace(/punct/g, gi).getRegex(), Ci = K(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, hi).replace(/punct/g, di).getRegex(), wi = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", Ti = K(wi, "gu").replace(/notPunctSpace/g, pi).replace(/punctSpace/g, fi).replace(/punct/g, di).getRegex(), Ei = K(wi, "gu").replace(/notPunctSpace/g, vi).replace(/punctSpace/g, _i).replace(/punct/g, gi).getRegex(), Di = K("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, pi).replace(/punctSpace/g, fi).replace(/punct/g, di).getRegex(), Oi = K("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, pi).replace(/punctSpace/g, fi).replace(/punct/g, di).getRegex(), ki = K("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, pi).replace(/punctSpace/g, fi).replace(/punct/g, di).getRegex(), Ai = K(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, di).getRegex(), ji = K("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, pi).replace(/punctSpace/g, fi).replace(/punct/g, di).getRegex(), Mi = K(/\\(punct)/, "gu").replace(/punct/g, di).getRegex(), Ni = K(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, di).getRegex(), Pi = K(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), Fi = K(Qr).replace("(?:-->|$)", "-->").getRegex(), Ii = K("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", Fi).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Li = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Ri = K(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Li).getRegex(), zi = K(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Ri).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), Bi = K(/^!?\[(label)\]\[(ref)\]/).replace("label", Ri).replace("ref", Jr).getRegex(), Vi = K(/^!?\[(ref)\](?:\[\])?/).replace("ref", Jr).getRegex(), Hi = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, Ui = K(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Li).getRegex(), Wi = K("reflink|nolink(?!\\()", "g").replace("reflink", K(/^!?\[(label)\]\[(ref)\]/).replace("label", Ui).replace("ref", Hi).getRegex()).replace("nolink", K(/^!?\[(ref)\](?:\[\])?/).replace("ref", Hi).getRegex()).getRegex(), Gi = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, Ki = K(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), qi = {
	_backpedal: Pr,
	anyPunctuation: Mi,
	anyPunctuationOutsideCode: Ni,
	autolink: Pi,
	blockSkip: yi,
	br: li,
	code: ci,
	del: Pr,
	delLDelim: Pr,
	delRDelim: Pr,
	emStrongLDelim: xi,
	emStrongRDelimAst: Ti,
	emStrongRDelimUnd: Oi,
	escape: si,
	link: zi,
	nolink: Vi,
	punctuation: mi,
	reflink: Bi,
	reflinkSearch: Wi,
	tag: Ii,
	text: ui,
	url: Pr
}, Ji = {
	...qi,
	emStrongLDelim: Ci,
	emStrongRDelimAst: Di,
	emStrongRDelimUnd: ki,
	link: K(/^!?\[(label)\]\((.*?)\)/).replace("label", Ri).getRegex(),
	reflink: K(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Ri).getRegex()
}, Yi = {
	...qi,
	emStrongRDelimAst: Ei,
	emStrongLDelim: Si,
	delLDelim: Ai,
	delRDelim: ji,
	url: K(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", Ki).replace("protocol", Gi).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: K(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", Gi).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", Ir ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, Xi = {
	...Yi,
	br: K(li).replace("{2,}", "*").getRegex(),
	text: K(Yi.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, Zi = {
	normal: ri,
	gfm: ai,
	pedantic: oi
}, Qi = {
	normal: qi,
	gfm: Yi,
	breaks: Xi,
	pedantic: Ji
}, $i = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, ea = (e) => $i[e];
function ta(e, t) {
	if (t) {
		if (q.escapeTest.test(e)) return e.replace(q.escapeReplace, ea);
	} else if (q.escapeTestNoEncode.test(e)) return e.replace(q.escapeReplaceNoEncode, ea);
	return e;
}
function na(e) {
	return q.numericCharacterReference.test(e) ? e.replace(q.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function ra(e) {
	try {
		e = encodeURI(e).replace(q.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function ia(e, t) {
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
function aa(e, t, n) {
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
function oa(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && q.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function sa(e) {
	return e.trim().replace(q.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function ca(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var la = /* @__PURE__ */ new WeakMap();
function ua(e) {
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
function da(e, t) {
	let n = ua(t), r = la.get(e);
	r ? r.push(n) : la.set(e, [n]);
}
function fa(e) {
	let t = la.get(e);
	t && (t.pop(), t.length === 0 && la.delete(e));
}
function pa(e, t, n) {
	let r = la.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : ua(t).lastPossibleEmStrongCloserIndex[n];
}
function ma(e, t) {
	let n = la.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : ua(t).lastPossibleDelCloserIndex;
}
function ha(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function ga(e, t, n, r, i) {
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
function _a(e, t, n) {
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
function va(e, t, n, r) {
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
var ya = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || Mr;
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
			let e = this.options.pedantic ? t[0] : oa(t[0]);
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
			let e = t[0], n = _a(e, t[3] || "", this.rules);
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
				let t = aa(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: aa(t[0], "\n"),
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
			raw: aa(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = aa(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + ha(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? ha(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => ha(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => ha(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = oa(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && ca(t[2], "()") !== -1) return;
			let e = sa(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: aa(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = ia(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: aa(t[0], "\n"),
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
			for (let e of i) a.rows.push(ia(e, a.header.length).map((e, t) => ({
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
				raw: aa(t[0], "\n"),
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
			if (!this.options.pedantic && va(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = aa(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = ca(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = aa(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), ga(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && va(e, n[1], r, this.rules)) return;
			let i = t[sa(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return ga(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = pa(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = ma(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : na(t[0]),
				escaped: e
			};
		}
	}
}, ba = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || Mr, this.options.tokenizer = this.options.tokenizer || new ya(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: q,
			block: Zi.normal,
			inline: Qi.normal
		};
		this.options.pedantic ? (t.block = Zi.pedantic, t.inline = Qi.pedantic) : this.options.gfm && (t.block = Zi.gfm, t.inline = this.options.breaks ? Qi.breaks : Qi.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: Zi,
			inline: Qi
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, sa(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== ya.prototype.emStrong || this.tokenizer.del !== ya.prototype.del ? this.#n(e) : e;
		da(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			fa(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, sa(t.slice(r + 1, -1)))) return t;
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
}, xa = class {
	options;
	parser;
	constructor(e) {
		this.options = e || Mr;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(q.notSpaceStart)?.[0], i = e ? e.replace(q.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + ta(r) + "\">" + (n ? i : ta(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : ta(i, !0)) + "</code></pre>\n";
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
		return `<code>${ta(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? ta(n, !0) : this.parser.parseInline(r), o = ra(e);
		if (o === null) return a;
		e = ta(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + ta(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = ra(e);
		if (i === null) return ta(n);
		e = i;
		let a = `<img src="${ta(e)}" alt="${ta(n)}"`;
		return t && (a += ` title="${ta(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : ta(e.text);
	}
}, Sa = class {
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
}, Ca = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || Mr, this.options.renderer = this.options.renderer || new xa(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Sa(), this.textRenderer.parser = this;
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
}, wa = class {
	options;
	block;
	constructor(e) {
		this.options = e || Mr;
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
		return e ? ba.lex : ba.lexInline;
	}
	provideParser(e = this.block) {
		return e ? Ca.parse : Ca.parseInline;
	}
};
function Ta(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var Ea = new class {
	defaults = jr();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = Ca;
	Renderer = xa;
	TextRenderer = Sa;
	Lexer = ba;
	Tokenizer = ya;
	Hooks = wa;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (Ta(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) Ta(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) Ta(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				Ta(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					Ta(n, this.walkTokens(i, t));
				}) : e.tokens && Ta(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new xa(this.defaults);
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
				let t = this.defaults.tokenizer || new ya(this.defaults);
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
				let t = this.defaults.hooks || new wa();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = wa.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && wa.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return Ta(n, r.call(this, e)), t && Ta(n, t.call(this, e)), n;
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
		return ba.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return Ca.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? ba.lex : ba.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? Ca.parse : Ca.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? ba.lex : ba.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? Ca.parse : Ca.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + ta(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function J(e, t) {
	return Ea.parse(e, t);
}
J.options = J.setOptions = function(e) {
	return Ea.setOptions(e), J.defaults = Ea.defaults, Nr(J.defaults), J;
}, J.getDefaults = jr, J.defaults = Mr;
function Da(...e) {
	return Ea.use(...e), J.defaults = Ea.defaults, Nr(J.defaults), J;
}
J.use = Da, J.walkTokens = function(e, t) {
	return Ea.walkTokens(e, t);
}, J.parseInline = Ea.parseInline, J.Parser = Ca, J.parser = Ca.parse, J.Renderer = xa, J.TextRenderer = Sa, J.Lexer = ba, J.lexer = ba.lex, J.Tokenizer = ya, J.Hooks = wa, J.parse = J, J.options, J.setOptions, J.walkTokens, J.parseInline, Ca.parse, ba.lex;
// node_modules/dompurify/dist/purify.es.mjs
function Oa(e, t) {
	this.v = e, this.k = t;
}
function ka(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function Aa(e) {
	if (Array.isArray(e)) return e;
}
function ja(e, t) {
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
function Ma() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function Na(e, t) {
	return Aa(e) || ja(e, t) || Pa(e, t) || Ma();
}
function Pa(e, t) {
	if (e) {
		if (typeof e == "string") return ka(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? ka(e, t) : void 0;
	}
}
function Fa(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof Oa;
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
Fa.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, Fa.prototype.next = function(e) {
	return this._invoke("next", e);
}, Fa.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, Fa.prototype.return = function(e) {
	return this._invoke("return", e);
};
var Ia = Object.entries, La = Object.setPrototypeOf, Ra = Object.isFrozen, za = Object.getPrototypeOf, Ba = Object.getOwnPropertyDescriptor, Y = Object.freeze, Va = Object.seal, Ha = Object.create, Ua = typeof Reflect < "u" && Reflect, Wa = Ua.apply, Ga = Ua.construct;
Y ||= function(e) {
	return e;
}, Va ||= function(e) {
	return e;
}, Wa ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, Ga ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var Ka = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var qa = Z(Array.prototype.lastIndexOf), Ja = Z(Array.prototype.pop), Ya = Z(Array.prototype.push), Xa = Z(Array.prototype.slice), Za = Z(Array.prototype.splice), Qa = Array.isArray, $a = go(Set), eo = Z(Set.prototype.add), to = Z(Set.prototype.has), no = Z(String.prototype.toLowerCase), ro = Z(String.prototype.toString), io = Z(String.prototype.match), ao = Z(String.prototype.replace), oo = Z(String.prototype.indexOf), so = Z(String.prototype.trim), co = Z(Number.prototype.toString), lo = Z(Boolean.prototype.toString), uo = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), fo = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), po = Z(Object.prototype.hasOwnProperty), mo = Z(Object.prototype.toString), X = Z(RegExp.prototype.test), ho = go(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return Wa(e, t, n);
	};
}
function go(e) {
	return function() {
		return Ga(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : no;
	if (La && La(e, null), !Qa(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Ra(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function _o(e) {
	for (let t = 0; t < e.length; t++) po(e, t) || (e[t] = null);
	return e;
}
function vo(e) {
	let t = Ha(null);
	for (let r of Ia(e)) {
		var n = Na(r, 2);
		let i = n[0], a = n[1];
		po(e, i) && (t[i] = Qa(a) ? _o(a) : a && typeof a == "object" && a.constructor === Object ? vo(a) : a);
	}
	return t;
}
function yo(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return co(e);
		case "boolean": return lo(e);
		case "bigint": return uo ? uo(e) : "0";
		case "symbol": return fo ? fo(e) : "Symbol()";
		case "undefined": return mo(e);
		case "function":
		case "object": {
			if (e === null) return mo(e);
			let t = e, n = bo(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : mo(e);
			}
			return mo(e);
		}
		default: return mo(e);
	}
}
function bo(e, t) {
	for (; e !== null;) {
		let n = Ba(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = za(e);
	}
	function n() {
		return null;
	}
	return n;
}
function xo(e, t) {
	for (; e !== null;) {
		let n = Ba(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = za(e);
	}
	return null;
}
function So(e) {
	try {
		return X(e, ""), !0;
	} catch {
		return !1;
	}
}
var Co = Y(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), wo = Y(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), To = Y([
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
]), Eo = Y([
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
]), Do = Y(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), Oo = Y([
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
]), ko = Y(["#text"]), Ao = Y(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), jo = Y(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), Mo = Y(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), No = Y([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), Po = Va(/{{[\w\W]*|^[\w\W]*}}/g), Fo = Va(/<%[\w\W]*|^[\w\W]*%>/g), Io = Va(/\${[\w\W]*/g), Lo = Va(/^data-[\-\w.\u00B7-\uFFFF]+$/), Ro = Va(/^aria-[\-\w]+$/), zo = Va(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), Bo = Va(/^(?:\w+script|data):/i), Vo = Va(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), Ho = Va(/^html$/i), Uo = Va(/^[a-z][.\w]*(-[.\w]+)+$/i), Wo = Va(/<[/\w!]/g), Go = Va(/<[/\w]/g), Ko = Va(/<\/no(script|embed|frames)/i), qo = Va(/\/>/i), $ = {
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
}, Jo = Y(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), Yo = function() {
	return typeof window > "u" ? null : window;
}, Xo = function(e, t) {
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
}, Zo = function() {
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
}, Qo = function(e, t, n, r) {
	return po(e, t) && Qa(e[t]) ? Q(r.base ? vo(r.base) : {}, e[t], r.transform) : n;
}, $o = function(e, t, n) {
	let r = po(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? vo(r) : n();
};
function es() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : Yo(), t = (e) => es(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = bo(d, "cloneNode"), p = bo(d, "remove"), m = bo(d, "removeAttributeNode"), h = bo(d, "nextSibling"), g = bo(d, "childNodes"), _ = bo(d, "parentNode"), v = bo(d, "shadowRoot"), y = bo(d, "attributes"), b = o && o.prototype ? bo(o.prototype, "nodeType") : null, x = o && o.prototype ? bo(o.prototype, "nodeName") : null, S = o && o.prototype ? bo(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = xo(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, ne = "", re, ie = !1, ae = 0, oe = function() {
		if (ae > 0) throw ho("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
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
		return ie ||= (re = Xo(u, i), !0), re;
	}, ue = n, de = ue.implementation, fe = ue.createNodeIterator, pe = ue.createDocumentFragment, E = ue.getElementsByTagName, me = r.importNode, D = Zo();
	t.isSupported = typeof Ia == "function" && typeof _ == "function" && de && de.createHTMLDocument !== void 0;
	let he = Po, ge = Fo, O = Io, _e = Lo, ve = Ro, ye = Bo, be = Vo, xe = Uo, Se = zo, k = null, Ce = Q({}, [
		...Co,
		...wo,
		...To,
		...Do,
		...ko
	]), A = null, we = Q({}, [
		...Ao,
		...jo,
		...Mo,
		...No
	]), Te = Object.seal(Ha(null, {
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
	})), Ee = null, De = null, Oe = Object.seal(Ha(null, {
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
	})), ke = !0, Ae = !0, j = !1, je = !0, Me = !1, Ne = !0, Pe = !1, Fe = !1, Ie = null, Le = null, Re = !1, ze = !1, M = !1, Be = !1, Ve = !0, He = !1, Ue = "user-content-", We = !0, Ge = !1, Ke = null, qe = {}, Je = null, Ye = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Xe = null, Ze = Q({}, [
		"audio",
		"video",
		"img",
		"source",
		"image",
		"track"
	]), Qe = null, $e = Q({}, [
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
	]), et = "http://www.w3.org/1998/Math/MathML", tt = "http://www.w3.org/2000/svg", nt = "http://www.w3.org/1999/xhtml", rt = nt, N = !1, it = null, at = Q({}, [
		et,
		tt,
		nt
	], ro), ot = Y([
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
		e && typeof e == "object" || (e = {}), e = vo(e), dt = ft.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, P = dt === "application/xhtml+xml" ? ro : no, k = Qo(e, "ALLOWED_TAGS", Ce, { transform: P }), A = Qo(e, "ALLOWED_ATTR", we, { transform: P }), it = Qo(e, "ALLOWED_NAMESPACES", at, { transform: ro }), Qe = Qo(e, "ADD_URI_SAFE_ATTR", $e, {
			transform: P,
			base: $e
		}), Xe = Qo(e, "ADD_DATA_URI_TAGS", Ze, {
			transform: P,
			base: Ze
		}), Je = Qo(e, "FORBID_CONTENTS", Ye, { transform: P }), Ee = Qo(e, "FORBID_TAGS", vo({}), { transform: P }), De = Qo(e, "FORBID_ATTR", vo({}), { transform: P }), qe = po(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? vo(e.USE_PROFILES) : e.USE_PROFILES : !1, ke = e.ALLOW_ARIA_ATTR !== !1, Ae = e.ALLOW_DATA_ATTR !== !1, j = e.ALLOW_UNKNOWN_PROTOCOLS || !1, je = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Me = e.SAFE_FOR_TEMPLATES || !1, Ne = e.SAFE_FOR_XML !== !1, Pe = e.WHOLE_DOCUMENT || !1, ze = e.RETURN_DOM || !1, M = e.RETURN_DOM_FRAGMENT || !1, Be = e.RETURN_TRUSTED_TYPE || !1, Re = e.FORCE_BODY || !1, Ve = e.SANITIZE_DOM !== !1, He = e.SANITIZE_NAMED_PROPS || !1, We = e.KEEP_CONTENT !== !1, Ge = e.IN_PLACE || !1, Se = So(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : zo, rt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : nt, st = $o(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, ot)), lt = $o(e, "HTML_INTEGRATION_POINTS", () => Q({}, ct));
		let t = $o(e, "CUSTOM_ELEMENT_HANDLING", () => Ha(null));
		if (Te = Ha(null), po(t, "tagNameCheck") && ht(t.tagNameCheck) && (Te.tagNameCheck = t.tagNameCheck), po(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (Te.attributeNameCheck = t.attributeNameCheck), po(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (Te.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), Va(Te), Me && (Ae = !1), M && (ze = !0), qe && (k = Q({}, ko), A = Ha(null), qe.html === !0 && (Q(k, Co), Q(A, Ao)), qe.svg === !0 && (Q(k, wo), Q(A, jo), Q(A, No)), qe.svgFilters === !0 && (Q(k, To), Q(A, jo), Q(A, No)), qe.mathMl === !0 && (Q(k, Do), Q(A, Mo), Q(A, No))), Oe.tagCheck = null, Oe.attributeCheck = null, po(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? Oe.tagCheck = e.ADD_TAGS : Qa(e.ADD_TAGS) && (k === Ce && (k = vo(k)), Q(k, e.ADD_TAGS, P))), po(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? Oe.attributeCheck = e.ADD_ATTR : Qa(e.ADD_ATTR) && (A === we && (A = vo(A)), Q(A, e.ADD_ATTR, P))), po(e, "ADD_FORBID_CONTENTS") && Qa(e.ADD_FORBID_CONTENTS) && (Je === Ye && (Je = vo(Je)), Q(Je, e.ADD_FORBID_CONTENTS, P)), We && (k["#text"] = !0), Pe && Q(k, [
			"html",
			"head",
			"body"
		]), k.table && (Q(k, ["tbody"]), delete Ee.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw ho("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw ho("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
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
		...wo,
		...To,
		...Eo
	]), vt = Q({}, [...Do, ...Oo]), yt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "svg" : t.namespaceURI === et ? e === "svg" && (n === "annotation-xml" || st[n]) : !!_t[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "math" : t.namespaceURI === tt ? e === "math" && lt[n] : !!vt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === tt && !lt[n] || t.namespaceURI === et && !st[n] ? !1 : !vt[e] && (ut[e] || !_t[e]);
	}, St = function(e) {
		let t = te(e), n = no(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = no(w(r))) : (i = { namespaceURI: rt }, a = "template"), it[t] ? t === tt ? yt(n, i, a) : t === et ? bt(n, i, a) : t === nt ? xt(n, i, a) : !!(dt === "application/xhtml+xml" && it[t]) : !1;
	}, F = function(e) {
		Ya(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw ho("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
		}
	}, Ct = function(e, t, n) {
		try {
			m(e, t);
		} catch {
			try {
				e.removeAttribute(n);
			} catch {}
		}
	}, wt = function(e) {
		Ot(e);
		let t = g(e);
		t && Ka(Xa(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && Ct(e, r, i);
		}
	}, Tt = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		Ya(t.removed, {
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
			if (ze || M) try {
				F(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Et = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || A[P(i)] || Ct(e, r, i);
		}
	}, Dt = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				L(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, Ot = function(e) {
		Dt(e, (e) => {
			C(e) === $.element && Et(e);
		});
	}, kt = function(e) {
		let t = [];
		return Dt(e, (e) => {
			t.push(e);
		}, !0), t;
	}, At = function(e, t) {
		let n = $a();
		Ka(kt(e), (e) => {
			C(e) === $.element && eo(n, e);
		}), Ka(t, (e) => {
			C(e) === $.element && !to(n, e) && Et(e);
		});
	}, jt = function(e, n) {
		let r = !1;
		return Ka(t.removed, (t) => {
			t.element && (r ||= t.element === e, Ot(t.element));
		}), At(e, n), r;
	}, Mt = function(e, t) {
		let n = Ke, r = null, i = () => (r || (r = $a(), Ka(t, (e) => {
			eo(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = kt(e), a = [e], o = !1;
			if (Ka(r, (e) => {
				C(e) === $.documentFragment ? Ya(a, e) : !o && !to(n, e) && (o = to(i(), e));
			}), !o) return;
			if (t >= 8) throw ho("the tree kept changing during in-place sanitization; refusing to return it");
			Ka(a, (t, n) => {
				Nt(t, n > 0, e, i());
			});
		}
	}, Nt = function(e, t, n, r) {
		let i = Rt(e), a;
		for (; a = i.nextNode();) to(r, a) && !to(Ke, a) && (t && Ht(D.uponSanitizeShadowNode, a, null), qt(a, n), en(a, n), L(a.content) && tn(a.content));
	}, Pt = function(e) {
		Ke && eo(Ke, e);
	}, Ft = function(e, t) {
		return Ne ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, It = function(e) {
		Ne && Dt(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && X(Go, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = P(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Ft("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, Lt = function(e) {
		let t = null, r = null;
		if (Re) e = "<remove></remove>" + e;
		else {
			let t = io(e, /^[\r\n\t ]+/);
			r = t && t[0];
		}
		dt === "application/xhtml+xml" && rt === nt && (e = "<html xmlns=\"http://www.w3.org/1999/xhtml\"><head></head><body>" + e + "</body></html>");
		let i = T ? se(e) : e;
		if (rt === nt) try {
			t = new l().parseFromString(i, dt);
		} catch {}
		if (!t || !t.documentElement) {
			t = de.createDocument(rt, "template", null);
			try {
				t.documentElement.innerHTML = N ? ne : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), rt === nt ? E.call(t, Pe ? "html" : "body")[0] : Pe ? t.documentElement : a;
	}, Rt = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return fe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, zt = function(e) {
		return e = ao(e, he, " "), e = ao(e, ge, " "), e = ao(e, O, " "), e;
	}, Bt = function(e) {
		I(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && Ka(t, (e) => {
			L(e.content) && Bt(e.content);
		});
	}, I = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = fe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = zt(r.data), r = n.nextNode();
	}, Vt = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || P(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, L = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, R = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function Ht(e, n, r) {
		e.length !== 0 && Ka(e, (e) => {
			e.call(t, n, r, pt);
		});
	}
	let Ut = function(e, t) {
		return !!(Ne && e.hasChildNodes() && !R(e.firstElementChild) && X(Wo, e.textContent) && X(Wo, e.innerHTML) || Ne && Jo[t] && C(e) === $.element && te(e) === nt && (R(e.firstElementChild) || typeof e.textContent == "string" && X(Wo, e.textContent)) || e.nodeType === $.processingInstruction || Ne && e.nodeType === $.comment && X(Go, e.data));
	}, Wt = function(e, t) {
		return e instanceof RegExp ? X(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Gt = function(e, t, n) {
		if (!Ee[t] && Xt(t) && Wt(Te.tagNameCheck, t)) return !1;
		if (We && !Je[t] && !Je[no(t)]) {
			let t = _(e), r = g(e);
			if (r && t) {
				let i = r.length;
				for (let a = i - 1; a >= 0; --a) {
					let i = e === n ? f(r[a], !0) : r[a];
					t.insertBefore(i, h(e));
				}
			}
		}
		return F(e), !0;
	}, Kt = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? vo(t) : t;
	}, z = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ge && Ot(e), !0);
	}, qt = function(e, n) {
		if (Pt(e), Ht(D.beforeSanitizeElements, e, null), z(e, n)) return !0;
		if (Vt(e)) return F(e), !0;
		let r = P(w(e));
		if (k = Kt(D.uponSanitizeElement, k, Ce, Ie), Ht(D.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: k
		}), z(e, n)) return !0;
		if (Ut(e, r)) return F(e), !0;
		if (Ee[r] || !(Oe.tagCheck instanceof Function && Oe.tagCheck(r)) && !k[r]) {
			let t = Gt(e, r, n);
			return t === !1 && (Ht(D.afterSanitizeElements, e, null), z(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !St(e) || (r === "noscript" || r === "noembed" || r === "noframes") && X(Ko, e.innerHTML)) return F(e), !0;
		if (Me && e.nodeType === $.text) {
			let n = zt(e.textContent);
			e.textContent !== n && (Ya(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return Ht(D.afterSanitizeElements, e, null), z(e, n);
	}, Jt = function(e, t, r) {
		if (De[t] || Ft(t, e) || Ve && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = A[t] || Oe.attributeCheck instanceof Function && Oe.attributeCheck(t, e);
		return Ae && X(_e, t) || ke && X(ve, t) ? !0 : i ? Qe[t] || X(Se, ao(r, be, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && oo(r, "data:") === 0 && Xe[e] || j && !X(ye, ao(r, be, "")) ? !0 : !r : Xt(e) && Wt(Te.tagNameCheck, e) && Wt(Te.attributeNameCheck, t, e) || t === "is" && Te.allowCustomizedBuiltInElements && Wt(Te.tagNameCheck, r);
	}, Yt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), Xt = function(e) {
		return !Yt[no(e)] && X(xe, e);
	}, Zt = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return se(r);
			case "TrustedScriptURL": return ce(r);
		}
		return r;
	}, Qt = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Vt(e) || (F(e), !1);
		} catch {
			return Tt(t, e), !1;
		}
	}, $t = function(e, t) {
		return Ne && X(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : no(e) === "attributename" && io(t, "href") !== null;
	}, en = function(e, n) {
		if (Ht(D.beforeSanitizeAttributes, e, null), z(e, n)) return;
		let r = e.attributes;
		if (!r || Vt(e)) return;
		A = Kt(D.uponSanitizeAttribute, A, we, Le);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: A,
			forceKeepAttr: void 0
		}, a = r.length, o = P(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = P(s), d = l, f = s === "value" ? d : so(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, Ht(D.uponSanitizeAttribute, e, i), f = i.attrValue, He && (u === "id" || u === "name") && oo(f, Ue) !== 0 && (Tt(s, e, n), f = Ue + f, p = !0), $t(s, f) ? Tt(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !je && X(qo, f) ? Tt(s, e, n) : (Me && (f = zt(f)), Jt(o, u, f) ? (f = Zt(o, u, c, f), f !== d && Qt(e, s, c, f) && p && Ja(t.removed)) : Tt(s, e, n)) : Tt(s, e, n));
		}
		Ht(D.afterSanitizeAttributes, e, null), z(e, n);
	}, tn = function(e) {
		let t = null, n = Rt(e);
		for (Ht(D.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (Ht(D.uponSanitizeShadowNode, t, null), qt(t, e), en(t, e), L(t.content) && tn(t.content), C(t) === $.element) {
			let e = v(t);
			L(e) && (nn(e), tn(e));
		}
		Me && I(e), Ht(D.afterSanitizeShadowDOM, e, null);
	}, nn = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				tn(e.shadow);
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
					L(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				L(e) && t.push({
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
		if (N = !e, N && (e = "<!-->"), typeof e != "string" && !R(e) && (e = yo(e), typeof e != "string")) throw ho("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Fe ? (k = Ie, A = Le) : gt(n), (D.uponSanitizeElement.length > 0 || D.uponSanitizeAttribute.length > 0) && (k = vo(k)), D.uponSanitizeAttribute.length > 0 && (A = vo(A)), t.removed = [];
		let c = Ge && typeof e != "string" && R(e), l = [], u = Ke;
		if (c) {
			It(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = P(t);
				if (!k[n] || Ee[n]) throw wt(e), ho("root node is forbidden and cannot be sanitized in-place");
			}
			if (Vt(e)) throw wt(e), ho("root node is clobbered and cannot be sanitized in-place");
			l = kt(e), Ke = $a();
		} else if (R(e)) i = Lt("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), nn(i);
		else {
			if (!ze && !Me && !Pe && e.indexOf("<") === -1) return T && Be ? se(e) : e;
			if (i = Lt(e), !i) return ze ? null : Be ? ne : "";
		}
		i && Re && F(i.firstChild);
		let d = c ? e : i;
		try {
			c && nn(e);
			let t = Rt(d);
			for (; o = t.nextNode();) qt(o, d), en(o, d), L(o.content) && tn(o.content);
			c && Mt(e, l);
		} catch (t) {
			throw c && (wt(e), jt(e, l)), t;
		} finally {
			Ke = u;
		}
		if (c) {
			if (jt(e, l)) throw ho("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Me && Bt(e), e;
		}
		if (ze) {
			if (Me && Bt(i), M) for (s = pe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (A.shadowroot || A.shadowrootmode) && (s = me.call(r, s, !0)), s;
		}
		let f = Pe ? i.outerHTML : i.innerHTML;
		return Pe && k["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && X(Ho, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Me && (f = zt(f)), T && Be ? se(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Fe = !0, Ie = k, Le = A;
	}, t.clearConfig = function() {
		pt = null, Fe = !1, Ie = null, Le = null, T = re, ne = "";
	}, t.isValidAttribute = function(e, t, n) {
		pt || gt({});
		let r = P(e), i = P(t);
		return typeof n == "string" && ($t(t, n) || !je && X(qo, n)) ? !1 : Jt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && po(D, e) && Ya(D[e], t);
	}, t.removeHook = function(e, t) {
		if (po(D, e)) {
			if (t !== void 0) {
				let n = qa(D[e], t);
				return n === -1 ? void 0 : Za(D[e], n, 1)[0];
			}
			return Ja(D[e]);
		}
	}, t.removeHooks = function(e) {
		po(D, e) && (D[e] = []);
	}, t.removeAllHooks = function() {
		D = Zo();
	}, t;
}
es();
// kit/format.js
function ts(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
// plugins/pools/Pools.svelte
var ns = /* @__PURE__ */ U("<div class=\"card err\"> </div>"), rs = /* @__PURE__ */ U("<span class=\"faint\"> </span>"), is = /* @__PURE__ */ U("<span class=\"pill warn gap svelte-1td2jxo\">drained</span>"), as = /* @__PURE__ */ U("<span class=\"pill accent\"> </span>"), os = /* @__PURE__ */ U("<span class=\"muted\">auto</span>"), ss = /* @__PURE__ */ U("<button class=\"svelte-1td2jxo\">Auto</button>"), cs = /* @__PURE__ */ U("<tr><td class=\"mono c-id\"> <!></td><td data-label=\"desired\"> </td><td data-label=\"open\"> </td><td data-label=\"working\"> </td><td data-label=\"idle\"> </td><td data-label=\"cordoned\"> </td><td class=\"muted\" data-label=\"min/max\"> </td><td class=\"c-st\"><!></td><td class=\"c-act acts svelte-1td2jxo\"><button class=\"svelte-1td2jxo\">Scale</button> <!></td></tr>"), ls = /* @__PURE__ */ U("<h2 class=\"st\">Scale<!></h2> <!> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"scale\"><thead><tr><th>Pool</th><th>Desired</th><th>Open</th><th>Working</th><th>Idle</th><th>Cordoned</th><th>Min/max</th><th>Target</th><th></th></tr></thead><tbody></tbody></table></div>", 1), us = /* @__PURE__ */ U("<tr><td class=\"mono c-id\"> </td><td class=\"c-st\"><span class=\"pill bad\">quarantined</span></td><td class=\"muted full\"> </td></tr>"), ds = /* @__PURE__ */ U("<tr><td class=\"mono c-id\"> </td><td class=\"c-st\"><span class=\"pill warn\">paused</span></td><td class=\"muted full\"> </td></tr>"), fs = /* @__PURE__ */ U("<tr><td class=\"faint\">All pools start sessions.</td></tr>"), ps = /* @__PURE__ */ U("<div class=\"faint\"> </div>"), ms = /* @__PURE__ */ U("<tr><td class=\"mono c-id\"> <!></td><td class=\"c-st\"><span> </span></td><td class=\"muted full\"> </td><td class=\"muted\" data-label=\"last fired\"> </td><td class=\"muted\" data-label=\"next cron\"> </td></tr>"), hs = /* @__PURE__ */ U("<tr><td class=\"faint\">No triggers.</td></tr>"), gs = /* @__PURE__ */ U("<!> <h2 class=\"st\">Pool status</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"pools\"><tbody><!><!><!></tbody></table></div> <h2 class=\"st\">Triggers</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"triggers\"><thead><tr><th>Trigger</th><th>Holds</th><th>Why not / stalls</th><th>Last fired</th><th>Next cron</th></tr></thead><tbody></tbody></table></div>", 1), _s = /* @__PURE__ */ U("<header class=\"vh\"><h1>Pools and triggers</h1></header> <div class=\"pad\"><!></div>", 1);
function vs(e, t) {
	Me(t, !0);
	let n = kr(() => t.api, () => "api/pools"), r = /* @__PURE__ */ F("");
	async function i(e) {
		wt(r, "");
		try {
			await t.api.post("api/pool/scale", e), n.refresh();
		} catch (e) {
			wt(r, Ar(e), !0);
		}
	}
	function a(e) {
		let t = window.prompt("Target sessions for " + e.name + " (" + e.min + ".." + e.max + ")", String(e.target ?? e.desired));
		if (t === null || t.trim() === "") return;
		let n = Number(t);
		!Number.isInteger(n) || n < 0 ? wt(r, "target must be a whole number") : i({
			pool: e.name,
			target: n
		});
	}
	var o = _s(), s = R(Vt(o), 2), c = I(s), l = (e) => {
		var t = ns(), r = L(t, !0);
		rn(() => G(r, n.error)), W(e, t);
	}, u = (e) => {
		let t = /* @__PURE__ */ Xe(() => Object.entries(n.data.quarantined ?? {})), o = /* @__PURE__ */ Xe(() => Object.entries(n.data.paused ?? {}));
		var s = gs(), c = Vt(s), l = (e) => {
			var t = ls(), o = Vt(t), s = R(I(o)), c = (e) => {
				var t = rs(), r = L(t);
				rn(() => G(r, `(host max ${n.data.host_max ?? ""})`)), W(e, t);
			};
			hr(s, (e) => {
				n.data.host_max && e(c);
			}), O(o);
			var l = R(o, 2), u = (e) => {
				var t = ns(), n = L(t, !0);
				rn(() => G(n, H(r))), W(e, t);
			};
			hr(l, (e) => {
				H(r) && e(u);
			});
			var d = R(l, 2), f = I(d), p = R(I(f));
			yr(p, 21, () => n.data.pools, (e) => e.name, (e, t) => {
				var n = cs(), r = I(n), o = I(r, !0), s = R(o), c = (e) => {
					W(e, is());
				};
				hr(s, (e) => {
					H(t).drained && e(c);
				}), O(r);
				var l = R(r), u = L(l, !0), d = R(l), f = L(d, !0), p = R(d), m = L(p, !0), h = R(p), g = L(h, !0), _ = R(h), v = L(_, !0), y = R(_), b = L(y), x = R(y), S = I(x), C = (e) => {
					var n = as(), r = L(n);
					rn(() => G(r, `target ${H(t).target ?? ""}`)), W(e, n);
				}, w = (e) => {
					W(e, os());
				};
				hr(S, (e) => {
					H(t).target == null ? e(w, -1) : e(C);
				}), O(x);
				var ee = R(x), te = I(ee), T = R(te, 2), ne = (e) => {
					var n = ss();
					Jn("click", n, () => i({
						pool: H(t).name,
						auto: !0
					})), W(e, n);
				};
				hr(T, (e) => {
					H(t).target != null && e(ne);
				}), O(ee), O(n), rn(() => {
					G(o, H(t).name), G(u, H(t).desired), G(f, H(t).open), G(m, H(t).working), G(g, H(t).idle), G(v, H(t).cordoned), G(b, `${H(t).min ?? ""}/${H(t).max ?? ""}`);
				}), Jn("click", te, () => a(H(t))), W(e, n);
			}), O(p), O(f), O(d), W(e, t);
		};
		hr(c, (e) => {
			n.data.pools?.length && e(l);
		});
		var u = R(c, 4), d = I(u), f = I(d), p = I(f);
		yr(p, 17, () => H(t), ([e, t]) => e, (e, t) => {
			var n = /* @__PURE__ */ Xe(() => h(H(t), 2));
			let r = () => H(n)[0], i = () => H(n)[1];
			var a = us(), o = I(a), s = L(o, !0), c = L(R(o, 2), !0);
			O(a), rn(() => {
				G(s, r()), G(c, i());
			}), W(e, a);
		});
		var m = R(p);
		yr(m, 17, () => H(o), ([e, t]) => e, (e, t) => {
			var n = /* @__PURE__ */ Xe(() => h(H(t), 2));
			let r = () => H(n)[0], i = () => H(n)[1];
			var a = ds(), o = I(a), s = L(o, !0), c = L(R(o, 2));
			O(a), rn((e) => {
				G(s, r()), G(c, `until ${e ?? ""}`);
			}, [() => ts(String(i()))]), W(e, a);
		});
		var g = R(m), _ = (e) => {
			W(e, fs());
		};
		hr(g, (e) => {
			H(t).length === 0 && H(o).length === 0 && e(_);
		}), O(f), O(d), O(u);
		var v = R(u, 4), y = I(v), b = R(I(y));
		yr(b, 21, () => n.data.triggers, (e) => e.name, (e, t) => {
			var n = ms(), r = I(n), i = I(r, !0), a = R(i), o = (e) => {
				var n = ps(), r = L(n);
				rn(() => G(r, `open ${H(t).open ?? ""}`)), W(e, n);
			};
			hr(a, (e) => {
				H(t).open && e(o);
			}), O(r);
			var s = R(r), c = I(s), l = L(c, !0);
			O(s);
			var u = R(s), d = L(u, !0), f = R(u), p = L(f, !0), m = L(R(f), !0);
			O(n), rn((e, n, r) => {
				G(i, H(t).name), Dr(c, 1, `pill ${H(t).holds ? "accent" : ""}`), G(l, H(t).holds ? "fires next" : "waiting"), G(d, e), G(p, n), G(m, r);
			}, [
				() => H(t).stalls?.length ? H(t).stalls.join(", ") : H(t).why ?? "",
				() => ts(H(t).last_fired),
				() => ts(H(t).next_cron)
			]), W(e, n);
		}, (e) => {
			W(e, hs());
		}), O(b), O(y), O(v), W(e, s);
	};
	hr(c, (e) => {
		n.error ? e(l) : n.data && e(u, 1);
	}), O(s), W(e, o), Ne();
}
// plugins/pools/index.js
Yn(["click"]), Or("desk-pools-view", vs);
