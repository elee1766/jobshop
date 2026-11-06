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
function _(e, t) {
	if (Array.isArray(e)) return e;
	if (t === void 0 || !(Symbol.iterator in e)) return Array.from(e);
	let n = [];
	for (let r of e) if (n.push(r), n.length === t) break;
	return n;
}
var v = 1024, y = 2048, b = 4096, x = 8192, S = 16384, ee = 32768, C = 1 << 25, te = 65536, w = 1 << 19, ne = 1 << 20, T = 1 << 25, re = 1 << 21, ie = 1 << 22, ae = 1 << 23, oe = Symbol("$state"), E = Symbol("component"), se = Symbol(""), ce = Symbol("attributes"), le = Symbol("class"), ue = Symbol("style"), de = Symbol("text"), fe = Symbol("form reset"), pe = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), me = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function he() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ge(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function _e() {
	console.warn("https://svelte.dev/e/select_multiple_invalid_value");
}
function ve() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var D = !1;
function ye(e) {
	D = e;
}
var O;
function k(t) {
	if (t === null) throw ge(), e;
	return O = t;
}
function be() {
	return k(/* @__PURE__ */ Yt(O));
}
function A(t) {
	if (D) {
		if (/* @__PURE__ */ Yt(O) !== null) throw ge(), e;
		O = t;
	}
}
function xe(e = 1) {
	if (D) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ Yt(n);
		O = n;
	}
}
function Se(e = !0) {
	for (var t = 0, n = O;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Yt(n);
		e && n.remove(), n = i;
	}
}
function Ce(t) {
	if (!t || t.nodeType !== 8) throw ge(), e;
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
var j = null;
function Le(e) {
	j = e;
}
function Re(e, t = !1, n) {
	j = {
		p: j,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: G,
		l: null
	};
}
function ze(e) {
	var t = j, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) un(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, j = t.p, Be(e);
}
function Be(e = {}) {
	return s(e, E, { value: !0 }), e;
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
	if (He.length === 0 && !mt) {
		var t = He;
		queueMicrotask(() => {
			t === He && Ue();
		});
	}
	He.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ge = ~(y | b | v);
function M(e, t) {
	e.f = e.f & Ge | t;
}
function Ke(e) {
	e.f & 512 || e.deps === null ? M(e, v) : M(e, b);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function qe(e, t) {
	e.f & 2048 ? t.set(e, y) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, b), M(e, v);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
var Je = !1;
function Ye() {
	Je || (Je = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[fe]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Xe(e) {
	var t = U, n = G;
	W(null), kn(null);
	try {
		return e();
	} finally {
		W(t), kn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ze(e, t, n, r) {
	let i = Ve() ? tt : it;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = G, c = Qe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				nn(e, s);
			}
			$e();
		}
	}
	var d = et();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ rt(e))).then(u).catch((e) => nn(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), $e());
	}) : f();
}
function Qe() {
	var e = G, t = U, n = j, r = P;
	return function(i = !0) {
		kn(e), W(t), Le(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function $e(e = !0) {
	kn(null), W(null), Le(null), e && P?.deactivate();
}
function et() {
	var e = G, t = e.b, n = P, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function tt(e) {
	var n = 2 | y;
	return G !== null && (G.f |= w), {
		ctx: j,
		deps: null,
		effects: null,
		equals: we,
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
var nt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function rt(e, n, r) {
	let i = G;
	i === null && De();
	var a = void 0, o = At(t), s = !U, c = /* @__PURE__ */ new Set();
	return fn(() => {
		var t = G, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== pe && n.reject(e);
			}).finally($e);
		} catch (e) {
			n.reject(e), $e();
		}
		var r = P;
		if (s) {
			if (t.f & 32768) var l = et();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(nt);
			else for (let e of c.values()) e.reject(nt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== nt && (r.activate(), t ? (o.f |= ae, Ft(o, t)) : (o.f & 8388608 && (o.f ^= ae), Ft(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), cn(() => {
		for (let e of c) e.reject(nt);
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
function N(e) {
	let t = /* @__PURE__ */ tt(e);
	return jn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function it(e) {
	let t = /* @__PURE__ */ tt(e);
	return t.equals = Ee, t;
}
function at(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) H(t[n]);
	}
}
function ot(e) {
	var n, r = G, i = e.parent;
	if (!En && i !== null && e.v !== t && i.f & 24576) return he(), e.v;
	kn(i);
	try {
		at(e), n = Bn(e);
	} finally {
		kn(r);
	}
	return n;
}
function st(e) {
	var t = ot(e);
	!e.equals(t) && (e.wv = Ln(), (!P?.is_fork || e.deps === null) && (P === null ? e.v = t : (P.capture(e, t, !0), dt?.capture(e, t, !0)), e.deps === null)) ? M(e, v) : En || (ft === null ? Ke(e) : (sn() || P?.is_fork) && ft.set(e, t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Xe(() => {
		t.ac.abort(pe), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Un(t, 0), gn(t));
}
function lt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Wn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ut = null, P = null, dt = null, ft = null, pt = null, mt = !1, ht = !1, gt = null, _t = null, vt = 0, yt = 1, bt = class e {
	id = yt++;
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
			for (var r of n.d) M(r, y), t(r);
			for (r of n.m) M(r, b), t(r);
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
					t.f ^= v;
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
		for (var t = gt = [], n = [], r = _t = []; this.#a.length > 0;) {
			vt++ > 1e3 && (this.#b(), xt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Et(e), this.#d() || this.discard(), t;
			}
		}
		if (P = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (gt = null, _t = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Tt(e, t);
			r.length > 0 && P.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), dt = this, Ct(n), Ct(t), dt = null, this.#i?.resolve();
			var o = P;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ot.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= v;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= v : i & 4 ? t.push(r) : Rn(r) && (i & 16 && this.add_dirty_reaction(r, b), Wn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, y), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), P = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) qe(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), ft?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		P = this;
	}
	deactivate() {
		P = null, ft = null;
	}
	flush() {
		try {
			ht = !0, P = this, this.#m();
		} finally {
			vt = 0, pt = null, gt = null, _t = null, ht = !1, P = null, ft = null, Ot.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(nt);
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
		if (P === null) {
			let t = P = new e();
			!ht && We(() => {
				t.started || t.flush();
			});
		}
		return P;
	}
	apply() {
		ft = null;
	}
	schedule(e) {
		pt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ut = e : t.prev = e, this.linked = !1;
		}
	}
};
function xt() {
	try {
		Me();
	} catch (e) {
		nn(e, pt);
	}
}
var St = null;
function Ct(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Rn(r) && (St = /* @__PURE__ */ new Set(), Wn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && yn(r), St?.size > 0)) {
				Ot.clear();
				for (let e of St) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) St.has(n) && (St.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Wn(n);
					}
				}
				St.clear();
			}
		}
		St = null;
	}
}
function wt(e) {
	P.schedule(e);
}
function Tt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, v);
		for (var n = e.first; n !== null;) Tt(n, t), n = n.next;
	}
}
function Et(e) {
	M(e, v);
	for (var t = e.first; t !== null;) Et(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Dt = /* @__PURE__ */ new Set(), Ot = /* @__PURE__ */ new Map(), kt = !1;
function At(e, t) {
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
function jt(e, t) {
	let n = At(e, t);
	return jn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Mt(e, t = !1, n = !0) {
	let r = At(e);
	return t || (r.equals = Ee), r;
}
function F(e, t, n = !1) {
	return U !== null && (!On || U.f & 131072) && Ve() && U.f & 4325394 && (An === null || !An.has(e)) && Fe(), Ft(e, n ? Bt(t) : t, _t);
}
var Nt = null, Pt = 0;
function Ft(e, t, n = null) {
	if (!e.equals(t)) {
		En ? Ot.set(e, t) : Ot.has(e) || Ot.set(e, e.v);
		var r = bt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ot(t), ft === null && Ke(t);
		}
		e.wv = Ln(), Nt = null, Pt = 0, zt(e, y, n), Nt = null, Ve() && G !== null && G.f & 1024 && !(G.f & 96) && (J === null ? Mn([e]) : J.push(e)), !r.is_fork && Dt.size > 0 && !kt && It();
	}
	return t;
}
function It() {
	kt = !1;
	for (let e of Dt) {
		e.f & 1024 && M(e, b);
		let t;
		try {
			t = Rn(e);
		} catch {
			t = !0;
		}
		t && Wn(e);
	}
	Dt.clear();
}
function Lt(e, t = 1) {
	var n = Y(e), r = t === 1 ? n++ : n--;
	return F(e, n), r;
}
function Rt(e) {
	F(e, e.v + 1);
}
function zt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ve(), a = r.length;
		if (Pt += a, Pt > 1e5 && Nt === null && (Nt = /* @__PURE__ */ new Set()), Nt !== null) {
			if (Nt.has(e)) return;
			Nt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== G) {
				var l = (c & y) === 0;
				if (l && M(s, t), c & 131072) Dt.add(s);
				else if (c & 2) {
					var u = s;
					ft?.delete(u), zt(u, b, n);
				} else if (l) {
					var d = s;
					c & 16 && St !== null && St.add(d), n === null ? wt(d) : n.push(d);
				}
			}
		}
	}
}
function Bt(e) {
	if (typeof e != "object" || !e || oe in e || E in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ jt(0), s = null, l = Fn, p = (e) => {
		if (Fn === l) return e();
		var t = U, n = Fn;
		W(null), In(l);
		var r = e();
		return W(t), In(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ jt(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ne();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ jt(n.value, s);
				return i.set(t, e), e;
			}) : F(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ jt(t, s));
					i.set(n, e), Rt(o);
				}
			} else F(r, t), Rt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === oe) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ jt(Bt(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = Y(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = Y(a);
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
			return (r !== void 0 || (G !== null || U !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ jt(a ? Bt(e[n]) : t, s)), i.set(n, r)), Y(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ jt(t, s)), i.set(f + "", m)) : F(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ jt(void 0, s)), F(u, Bt(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => Bt(r));
				F(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && F(_, v + 1);
				}
				Rt(o);
			}
			return !0;
		},
		ownKeys(e) {
			Y(o);
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
function Vt(e) {
	try {
		if (typeof e == "object" && e && oe in e) return e[oe];
	} catch {}
	return e;
}
function Ht(e, t) {
	return Object.is(Vt(e), Vt(t));
}
var Ut, Wt, Gt, Kt;
function qt() {
	if (Ut === void 0) {
		Ut = window, Wt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Gt = c(t, "firstChild").get, Kt = c(t, "nextSibling").get, p(e) && (e[le] = void 0, e[ce] = null, e[ue] = void 0, e.__e = void 0), p(n) && (n[de] = void 0);
	}
}
function I(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Jt(e) {
	return Gt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Yt(e) {
	return Kt.call(e);
}
function L(e, t) {
	if (!D) return /* @__PURE__ */ Jt(e);
	var n = /* @__PURE__ */ Jt(O);
	if (n === null) n = O.appendChild(I());
	else if (t && n.nodeType !== 3) {
		var r = I();
		return n?.before(r), k(r), r;
	}
	return t && en(n), k(n), n;
}
function Xt(e, t = !1) {
	if (!D) {
		var n = /* @__PURE__ */ Jt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Yt(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = I();
			return O?.before(r), k(r), r;
		}
		en(O);
	}
	return O;
}
function R(e, t = !1) {
	if (!D) return /* @__PURE__ */ Jt(e);
	var n = L(e, t);
	return A(e), n;
}
function z(e, t = 1, n = !1) {
	let r = D ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Yt(r);
	if (!D) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = I();
			return r === null ? i?.after(a) : r.before(a), k(a), a;
		}
		en(r);
	}
	return k(r), r;
}
function Zt(e) {
	e.textContent = "";
}
function Qt() {
	return !1;
}
function $t(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function en(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function tn(e) {
	var t = G;
	if (t === null) return U.f |= ae, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	nn(e, t);
}
function nn(t, n) {
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
function rn(e) {
	G === null && (U === null && je(e), Ae()), En && ke(e);
}
function an(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function on(e, t) {
	var n = G;
	n !== null && n.f & 8192 && (e |= x);
	var r = {
		ctx: j,
		deps: null,
		nodes: null,
		f: e | y | 512,
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
	if (e & 4) gt === null ? bt.ensure().schedule(r) : gt.push(r);
	else if (t !== null) {
		try {
			Wn(r);
		} catch (e) {
			throw H(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= te));
	}
	if (i !== null && (i.parent = n, n !== null && an(i, n), U !== null && U.f & 2 && !(e & 64))) {
		var a = U;
		(a.effects ??= []).push(i);
	}
	return r;
}
function sn() {
	return U !== null && !On;
}
function cn(e) {
	let t = on(8, null);
	return M(t, v), t.teardown = e, t;
}
function ln(e) {
	rn("$effect");
	var t = G.f;
	if (!U && t & 32 && j !== null && !j.i) {
		var n = j;
		(n.e ??= []).push(e);
	} else return un(e);
}
function un(e) {
	return on(4 | ne, e);
}
function dn(e) {
	bt.ensure();
	let t = on(64 | w, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? bn(t, () => {
			H(t), n(void 0);
		}) : (H(t), n(void 0));
	});
}
function fn(e) {
	return on(ie | w, e);
}
function pn(e, t = 0) {
	return on(8 | t, e);
}
function B(e, t = [], n = [], r = []) {
	Ze(r, t, n, (t) => {
		on(8, () => {
			e(...t.map(Y));
		});
	});
}
function mn(e, t = 0) {
	return on(16 | t, e);
}
function V(e) {
	return on(32 | w, e);
}
function hn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = En, r = U;
		Dn(!0), W(null);
		try {
			t.call(null);
		} catch (t) {
			nn(t, e.parent);
		} finally {
			Dn(n), W(r);
		}
	}
}
function gn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Xe(() => {
			e.abort(pe);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : H(n, t), n = r;
	}
}
function _n(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || H(t), t = n;
	}
}
function H(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (vn(e.nodes.start, e.nodes.end), n = !0), e.f |= C, gn(e, t && !n), Un(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	hn(e), e.f ^= C, e.f |= S;
	var i = e.parent;
	i !== null && i.first !== null && yn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function vn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Yt(e);
		e.remove(), e = n;
	}
}
function yn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function bn(e, t, n = !0) {
	var r = [];
	e.f |= 256, xn(e, r, !0);
	var i = () => {
		n && H(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function xn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= x;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				xn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Sn(e) {
	e.f &= -257, Cn(e, !0);
}
function Cn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= x, e.f & 1024 || (M(e, y), bt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			Cn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function wn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Yt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var Tn = null, En = !1;
function Dn(e) {
	En = e;
}
var U = null, On = !1;
function W(e) {
	U = e;
}
var G = null;
function kn(e) {
	G = e;
}
var An = null;
function jn(e) {
	U !== null && (U.f & 2097152 || U.f & 2) && (An ??= /* @__PURE__ */ new Set()).add(e);
}
var K = null, q = 0, J = null;
function Mn(e) {
	J = e;
}
var Nn = 1, Pn = 0, Fn = Pn;
function In(e) {
	Fn = e;
}
function Ln() {
	return ++Nn;
}
function Rn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Rn(a) && st(a), a.wv > e.wv) return !0;
		}
		t & 512 && ft === null && M(e, v);
	}
	return !1;
}
function zn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(An !== null && An.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? zn(a, t, !1) : t === a && (n ? M(a, y) : a.f & 1024 && M(a, b), wt(a));
	}
}
function Bn(e) {
	var t = K, n = q, r = J, i = U, a = An, o = j, s = On, c = Fn, l = e.f;
	K = null, q = 0, J = null, U = l & 96 ? null : e, An = null, Le(e.ctx), On = !1, Fn = ++Pn, e.ac !== null && (Xe(() => {
		e.ac.abort(pe);
	}), e.ac = null);
	try {
		e.f |= re;
		var u = e.fn, d = u();
		e.f |= ee;
		var f = Vn(e);
		if (Ve() && J !== null && !On && f !== null && !(e.f & 6146)) for (var p = 0; p < J.length; p++) zn(J[p], e);
		if (i !== null && i !== e) {
			if (Pn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Pn;
			if (t !== null) for (let e of t) e.rv = Pn;
			J !== null && (r === null ? r = J : r.push(...J));
		}
		return e.f & 8388608 && (e.f ^= ae), d;
	} catch (t) {
		return Vn(e), tn(t);
	} finally {
		e.f ^= re, K = t, q = n, J = r, U = i, An = a, Le(o), On = s, Fn = c;
	}
}
function Vn(e) {
	var t = e.deps, n = P?.is_fork;
	if (K !== null) {
		var r;
		if (n || Un(e, q), t !== null && q > 0) for (t.length = q + K.length, r = 0; r < K.length; r++) t[q + r] = K[r];
		else e.deps = t = K;
		if (sn() && e.f & 512) for (r = q; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && q < t.length && (Un(e, q), t.length = q);
	return t;
}
function Hn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (K === null || !a.call(K, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ke(c), c.ac !== null && Xe(() => {
			c.ac.abort(pe), c.ac = null, M(c, y);
		}), ct(c), Un(c, 0);
	}
}
function Un(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Hn(e, n[r]);
}
function Wn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		M(e, v);
		var n = G;
		G = e;
		try {
			t & 16777232 ? _n(e) : gn(e), hn(e);
			var r = Bn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Nn;
		} finally {
			G = n;
		}
	}
}
function Y(e) {
	var t = !!(e.f & 2);
	if (Tn?.add(e), U !== null && !On && !(G !== null && G.f & 16384) && (An === null || !An.has(e))) {
		var n = U.deps;
		if (U.f & 2097152) e.rv < Pn && (e.rv = Pn, K === null && n !== null && n[q] === e ? q++ : K === null ? K = [e] : K.push(e));
		else {
			U.deps ??= [], a.call(U.deps, e) || U.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [U] : a.call(r, U) || r.push(U);
		}
	}
	if (En && Ot.has(e)) return Ot.get(e);
	if (t) {
		var i = e;
		if (En) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Kn(i)) && (o = ot(i)), Ot.set(i, o), o;
		}
		var s = !(i.f & 512) && !On && U !== null && !!(U.f & 512), c = (i.f & ee) === 0;
		Rn(i) && (s && (i.f |= 512), st(i)), s && !c && (lt(i), Gn(i));
	}
	if (ft?.has(e)) return ft.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Gn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (lt(n), Gn(n));
	}
}
function Kn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ot.has(t) || t.f & 2 && Kn(t)) return !0;
	return !1;
}
function qn(e) {
	var t = On;
	try {
		return On = !0, e();
	} finally {
		On = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Jn = Symbol("events"), Yn = /* @__PURE__ */ new Set(), Xn = /* @__PURE__ */ new Set();
function Zn(e, t, n) {
	(t[Jn] ??= {})[e] = n;
}
function Qn(e) {
	for (var t = 0; t < e.length; t++) Yn.add(e[t]);
	for (var n of Xn) n(e);
}
var $n = null, er = !1;
function tr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	$n = e, er || (er = !0, setTimeout(() => {
		er = !1, $n = null;
	}));
	var o = 0, c = $n === e && e[Jn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Jn] = t;
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
		var d = U, f = G;
		W(null), kn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Jn]?.[r];
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
			e[Jn] = t, delete e.currentTarget, W(d), kn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var nr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function rr(e) {
	return nr?.createHTML(e) ?? e;
}
function ir(e) {
	var t = $t("template");
	return t.innerHTML = rr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function ar(e, t) {
	var n = G;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function X(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (D) return ar(O, null), O;
		i === void 0 && (i = ir(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Jt(i)));
		var t = r || Wt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Jt(t), s = t.lastChild;
			ar(o, s);
		} else ar(t, t);
		return t;
	};
}
function or(e = "") {
	if (!D) {
		var t = I(e + "");
		return ar(t, t), t;
	}
	var n = O;
	return n.nodeType === 3 ? en(n) : (n.before(n = I()), k(n)), ar(n, n), n;
}
function Z(e, t) {
	if (D) {
		var n = G;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), be();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var sr = ["touchstart", "touchmove"];
function cr(e) {
	return sr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function lr(e) {
	let t = 0, n = At(0), r;
	return () => {
		sn() && (Y(n), pn(() => (t === 0 && (r = qn(() => e(() => Rt(n)))), t += 1, () => {
			We(() => {
				--t, t === 0 && (r?.(), r = void 0, Rt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var ur = te | w;
function dr(e, t, n, r) {
	new fr(e, t, n, r);
}
var fr = class {
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
	#m = lr(() => (this.#p = At(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = G;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = G.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = mn(() => {
			if (D) {
				let e = Ce(this.#t);
				be();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, ur), D && (this.#e = O);
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
			this.#v() || (t ? ve() : (t = !0, n && Ie(), this.#s !== null && bn(this.#s, () => {
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
					nn(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (S | C)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = V(() => e(this.#e)), We(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = I(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return V(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							nn(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(P)) : this.#u === 0 && (this.#e.before(e), this.#c = null, bn(this.#o, () => {
					this.#o = null;
				}), this.#x(P));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = V(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				wn(this.#a, e);
				let t = this.#n.pending;
				this.#o = V(() => t(this.#e));
			} else this.#x(P);
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
		var t = G, n = U, r = j;
		kn(this.#i), W(this.#i), Le(this.#i.ctx);
		try {
			return bt.ensure(), e();
		} finally {
			kn(t), W(n), Le(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && bn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, We(() => {
			this.#d = !1, this.#p && Ft(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Y(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		P?.is_fork ? (this.#a && P.skip_effect(this.#a), this.#o && P.skip_effect(this.#o), this.#s && P.skip_effect(this.#s), P.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (H(this.#a), null), this.#o &&= (H(this.#o), null), this.#s &&= (H(this.#s), null), D && (k(this.#t), xe(), k(Se()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return V(() => {
						var r = G;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return nn(e, this.#i.parent), null;
				}
			}));
		};
		We(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					nn(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => nn(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[de] ??= e.nodeValue) && (e[de] = n, e.nodeValue = `${n}`);
}
function pr(e, t) {
	return hr(e, t);
}
var mr = /* @__PURE__ */ new Map();
function hr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	qt();
	var u = void 0, d = dn(() => {
		var c = r ?? n.appendChild(I());
		dr(c, { pending: () => {} }, (n) => {
			Re({});
			var r = j;
			if (s && (r.c = s), a && (i.$$events = a), D && ar(n, null), u = t(n, i) || Be(), D && (G.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw ge(), e;
			ze();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = cr(r);
					for (let e of [n, document]) {
						var a = mr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), mr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, tr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Yn)), Xn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = mr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, tr), t.delete(e), t.size === 0 && mr.delete(r)) : t.set(e, i);
			}
			Xn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return gr.set(u, d), u;
}
var gr = /* @__PURE__ */ new WeakMap();
function _r(e, t) {
	let n = gr.get(e);
	return n ? (gr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var vr = class {
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
			if (n) Sn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Sn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
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
						wn(r, t), t.append(I()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else H(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), bn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (H(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = P, r = Qt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = I();
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
		} else D && (this.anchor = O), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function $(e, t, n = !1) {
	var r;
	D && (r = O, be());
	var i = new vr(e), a = n ? te : 0;
	function o(e, t) {
		if (D) {
			var n = Ce(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Se();
				k(a), i.anchor = a, ye(!1), i.ensure(e, t), ye(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	mn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function yr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		bn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					br(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Zt(d), d.append(u), e.items.clear();
		}
		br(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function br(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= T, wn(a, document.createDocumentFragment())) : H(t[i], n);
	}
}
var xr;
function Sr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = D ? k(/* @__PURE__ */ Jt(u)) : u.appendChild(I());
	}
	D && be();
	var d = null, f = /* @__PURE__ */ it(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Y(f);
			_.fallback = d, wr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= T, Er(d, null, c)) : Sn(d) : bn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: mn(() => {
			var e = Y(f), r = e.length;
			let o = !1;
			D && Ce(c) === "[!" != (r === 0) && (c = Se(), k(c), ye(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = P, v = Qt(), y = 0; y < r; y += 1) {
				D && O.nodeType === 8 && O.data === "]" && (c = O, o = !0, ye(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Ft(S.v, b), S.i && Ft(S.i, y), v && _.unskip_effect(S.e)) : (S = Tr(l, m ? c : xr ??= I(), b, x, y, a, t, n), m || (S.e.f |= T), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = V(() => s(c)) : (d = V(() => s(xr ??= I())), d.f |= T)), r > u.size && Oe("", "", ""), D && r > 0 && k(Se()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && ye(!0), Y(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, D && (c = O);
}
function Cr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function wr(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = Cr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (Sn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= T, _ === l) Er(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Dr(e, d, _), Dr(e, _, y), Er(_, y, n), d = _, p = [], m = [], l = Cr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Er(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Dr(e, S.prev, ee.next), Dr(e, d, S), Dr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), Er(_, l, n), Dr(e, _.prev, _.next), Dr(e, _, d === null ? e.effect.first : d.next), Dr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Cr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Cr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (br(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = Cr(l.next);
		var te = C.length;
		if (te > 0) {
			var w = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.fix();
			}
			yr(e, C, w);
		}
	}
	a && We(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Tr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? At(n) : /* @__PURE__ */ Mt(n, !1, !1) : null, l = o & 2 ? At(i) : null;
	return {
		v: c,
		i: l,
		e: V(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Er(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Yt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Dr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var Or = [..." 	\n\r\f\xA0\v﻿"];
function kr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Or.includes(r[o - 1])) && (s === r.length || Or.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Ar(e, t, n, r, i, a) {
	var o = e[le];
	if (D || o !== n || o === void 0) {
		var s = kr(n, r, a);
		D && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[le] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/select.js
function jr(e, t) {
	t ? e.hasAttribute("selected") || e.setAttribute("selected", "") : e.removeAttribute("selected");
}
function Mr(e, t) {
	var n = e.__defaultValue, i = e.multiple, a = i ? n ?? [] : null;
	if (!i || r(a)) {
		var o = e.selectedIndex, s = t && i ? new Set(e.selectedOptions) : null;
		for (var c of e.options) {
			var l = Fr(c);
			jr(c, i ? a.includes(l) : Ht(l, n));
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
function Nr(e, t, n = !1) {
	if (e.multiple) {
		if (t == null) return;
		if (!r(t)) return _e();
		for (var i of e.options) i.selected = t.includes(Fr(i));
	} else {
		for (i of e.options) if (Ht(Fr(i), t)) {
			i.selected = !0;
			return;
		}
		n && t === void 0 || (e.selectedIndex = -1);
	}
}
function Pr(e) {
	var t = new MutationObserver((t) => {
		t.every(Ir) || ("__defaultValue" in e && Mr(e, !1), "__value" in e && Nr(e, e.__value));
	});
	t.observe(e, {
		childList: !0,
		subtree: !0,
		attributes: !0,
		attributeFilter: ["value"]
	}), cn(() => {
		t.disconnect();
	});
}
function Fr(e) {
	return "__value" in e ? e.__value : e.value;
}
function Ir(e) {
	if (e.target.closest("selectedcontent") !== null) return !0;
	if (e.type === "childList") {
		var t = [...e.addedNodes, ...e.removedNodes];
		return t.length > 0 && t.every((e) => e.nodeName === "SELECTEDCONTENT");
	}
	return !1;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Lr = Symbol("is custom element"), Rr = Symbol("is html"), zr = me ? "link" : "LINK", Br = me ? "progress" : "PROGRESS";
function Vr(e) {
	if (D) {
		var t = !1, n = () => {
			if (!t) {
				if (t = !0, e.hasAttribute("value")) {
					var n = e.value;
					Ur(e, "value", null), e.value = n;
				}
				if (e.hasAttribute("checked")) {
					var r = e.checked;
					Ur(e, "checked", null), e.checked = r;
				}
			}
		};
		e[fe] = n, We(n), Ye();
	}
}
function Hr(e, t) {
	var n = Wr(e);
	n.value !== (n.value = t ?? void 0) && (e.value !== t || t === 0 && e.nodeName === Br) && (e.value = t ?? "");
}
function Ur(e, t, n, r) {
	var i = Wr(e);
	if (!(D && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === zr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[se] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Kr(e).has(t) ? e[t] = n : (!D || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Wr(e) {
	return e[ce] ??= {
		[Lr]: e.nodeName.includes("-"),
		[Rr]: e.namespaceURI === n
	};
}
var Gr = /* @__PURE__ */ new Map();
function Kr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = Gr.get(t);
	if (n) return n;
	Gr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// kit/plugin.svelte.js
function qr(e, t) {
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
		#e = /* @__PURE__ */ jt(Bt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return Y(this.#e);
		}
		set props(e) {
			F(this.#e, e, !0);
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
			this.app ||= pr(t, {
				target: this,
				props: this.props
			}), i || (this.style.visibility = "hidden", e.then(() => {
				this.style.visibility = "";
			}));
		}
		disconnectedCallback() {
			this.app && _r(this.app), this.app = null;
		}
	});
}
function Jr(e, t) {
	let n = Bt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			Lt(r);
		}
	}), r = /* @__PURE__ */ jt(0);
	ln(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, Lt(r);
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
	return ln(() => {
		let n = e(), c = t();
		Y(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function Yr(e) {
	let t = Bt({ ...e()?.params?.() ?? {} });
	return ln(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
var Xr = (e) => String(e?.message || e);
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/format.js
function Zr(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
// plugins/pools/groups.js
var Qr = "(none)";
function $r(e) {
	let t = e?.quarantined ?? {}, n = e?.paused ?? {}, r = /* @__PURE__ */ new Map();
	for (let t of e?.pools ?? []) r.set(t.name, { ...t });
	for (let e of /* @__PURE__ */ new Set([...Object.keys(t), ...Object.keys(n)])) r.has(e) || r.set(e, {
		name: e,
		desired: 0,
		open: 0,
		working: 0,
		idle: 0,
		cordoned: 0,
		min: 0,
		max: 0
	});
	for (let [e, n] of Object.entries(t)) Object.assign(r.get(e), {
		quarantined: !0,
		why: n
	});
	for (let [e, t] of Object.entries(n)) Object.assign(r.get(e), {
		paused: !0,
		until: t
	});
	return [...r.values()].sort((e, t) => e.name.localeCompare(t.name));
}
function ei(e) {
	return !e.quarantined && !e.paused && !e.drained && e.target == null && !e.desired && !e.open && !e.working && !e.cordoned;
}
function ti(e) {
	return !e.holds && !e.stalls?.length;
}
function ni(e, t) {
	if (t === "none") return "";
	if (t === "project") return e.project || "(none)";
	let n = t + ":", r = (e.labels ?? []).find((e) => e.startsWith(n));
	return r ? r.slice(n.length) : Qr;
}
function ri(e) {
	let t = /* @__PURE__ */ new Set(["instance", "pack"]);
	for (let n of e) for (let e of n.labels ?? []) {
		let n = e.indexOf(":");
		n > 0 && t.add(e.slice(0, n));
	}
	return [
		"none",
		"project",
		...t
	];
}
function ii(e, t, n) {
	let r = t.trim().toLowerCase();
	return e.filter((e) => (!r || e.name.toLowerCase().includes(r)) && n.every((t) => (e.labels ?? []).includes(t)));
}
function ai(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) for (let e of n.labels ?? []) t.set(e, (t.get(e) ?? 0) + 1);
	return [...t.entries()].sort((e, t) => t[1] - e[1] || e[0].localeCompare(t[0]));
}
function oi(e, t) {
	let n = /* @__PURE__ */ new Map();
	for (let r of e) {
		let e = ni(r, t);
		n.set(e, [...n.get(e) ?? [], r]);
	}
	let r = [...n.entries()].map(([e, t]) => {
		let n = (e) => t.reduce((t, n) => t + e(n), 0);
		return {
			name: e,
			pools: t,
			desired: n((e) => e.desired),
			open: n((e) => e.open),
			working: n((e) => e.working),
			idle: n((e) => e.idle),
			quarantined: n((e) => +!!e.quarantined),
			paused: n((e) => +!!e.paused),
			drained: n((e) => +!!e.drained),
			targeted: n((e) => e.target == null ? 0 : 1),
			quiet: t.every(ei),
			shown: t.filter((e) => !ei(e)),
			folded: t.filter(ei)
		};
	}), i = (e) => e.quarantined || e.paused ? 0 : e.working || e.open ? 1 : e.quiet ? 3 : 2;
	return r.sort((e, t) => i(e) - i(t) || e.name.localeCompare(t.name));
}
function si(e, t) {
	let n = t === "project" ? "instance" : t, r = /* @__PURE__ */ new Map();
	for (let t of e) {
		let e = ni(t, n);
		r.set(e, [...r.get(e) ?? [], t]);
	}
	let i = (e) => e.stalls?.length ? 0 : e.holds ? 1 : 2;
	return [...r.entries()].map(([e, t]) => {
		let n = [...t].sort((e, t) => i(e) - i(t) || e.name.localeCompare(t.name));
		return {
			name: e,
			triggers: n,
			firing: t.filter((e) => e.holds).length,
			stalled: t.filter((e) => e.stalls?.length).length,
			quiet: t.every(ti),
			shown: n.filter((e) => !ti(e)),
			folded: n.filter(ti)
		};
	}).sort((e, t) => (e.quiet === t.quiet ? e.name.localeCompare(t.name) : e.quiet ? 1 : -1) || t.stalled - e.stalled);
}
// plugins/pools/Pools.svelte
var ci = (e, t = m) => {
	var n = vi(), r = L(n), i = L(r, !0), a = z(i), o = (e) => {
		var n = _i(), r = R(n);
		B(() => Q(r, `open ${t().open ?? ""}`)), Z(e, n);
	};
	$(a, (e) => {
		t().open && e(o);
	}), A(r);
	var s = z(r), c = L(s), l = R(c, !0);
	A(s);
	var u = z(s), d = R(u, !0), f = z(u), p = R(f, !0), h = R(z(f), !0);
	A(n), B((e, r, a) => {
		Ur(n, "data-trigger", t().name), Q(i, t().name), Ar(c, 1, `pill ${t().stalls?.length ? "bad" : t().holds ? "accent" : ""}`), Q(l, t().stalls?.length ? "stalled" : t().holds ? "fires next" : "waiting"), Q(d, e), Q(p, r), Q(h, a);
	}, [
		() => t().stalls?.length ? t().stalls.join(", ") : t().why ?? "",
		() => Zr(t().last_fired),
		() => Zr(t().next_cron)
	]), Z(e, n);
}, li = /* @__PURE__ */ X("<span class=\"pill bad gap svelte-1td2jxo\">quarantined</span>"), ui = /* @__PURE__ */ X("<span class=\"pill warn gap svelte-1td2jxo\">paused</span>"), di = /* @__PURE__ */ X("<span class=\"pill warn gap svelte-1td2jxo\">drained</span>"), fi = /* @__PURE__ */ X("<div class=\"muted why svelte-1td2jxo\"> </div>"), pi = /* @__PURE__ */ X("<span class=\"pill accent\"> </span>"), mi = /* @__PURE__ */ X("<span class=\"muted\">auto</span>"), hi = /* @__PURE__ */ X("<button class=\"svelte-1td2jxo\">Auto</button>"), gi = /* @__PURE__ */ X("<tr data-testid=\"pool-row\"><td class=\"mono c-id ind svelte-1td2jxo\"> <!> <!> <!> <!> <!></td><td data-label=\"desired\"> </td><td data-label=\"open\"> </td><td data-label=\"working\"> </td><td data-label=\"idle\"> </td><td data-label=\"cordoned\"> </td><td class=\"muted\" data-label=\"min/max\"> </td><td class=\"c-st\"><!></td><td class=\"c-act acts svelte-1td2jxo\"><button class=\"svelte-1td2jxo\">Scale</button> <!></td></tr>"), _i = /* @__PURE__ */ X("<div class=\"faint\"> </div>"), vi = /* @__PURE__ */ X("<tr data-testid=\"trigger-row\"><td class=\"mono c-id ind svelte-1td2jxo\"> <!></td><td class=\"c-st\"><span> </span></td><td class=\"muted full\"> </td><td class=\"muted\" data-label=\"last fired\"> </td><td class=\"muted\" data-label=\"next cron\"> </td></tr>"), yi = /* @__PURE__ */ X("<span class=\"sub\"> <!></span>"), bi = /* @__PURE__ */ X("<option> </option>"), xi = /* @__PURE__ */ X("<div class=\"card err\"> </div>"), Si = /* @__PURE__ */ X("<button data-testid=\"clear-filters\">Clear</button>"), Ci = /* @__PURE__ */ X("<button> <span class=\"faint\"> </span></button>"), wi = /* @__PURE__ */ X("<div class=\"chips lbl svelte-1td2jxo\" role=\"group\" aria-label=\"label filter\" data-testid=\"label-chips\"></div>"), Ti = /* @__PURE__ */ X("<span class=\"faint gap svelte-1td2jxo\" data-testid=\"idle-note\"> </span>"), Ei = /* @__PURE__ */ X("<span class=\"pill bad gap svelte-1td2jxo\"> </span>"), Di = /* @__PURE__ */ X("<span class=\"pill warn gap svelte-1td2jxo\"> </span>"), Oi = /* @__PURE__ */ X("<span class=\"pill accent gap svelte-1td2jxo\"> </span>"), ki = /* @__PURE__ */ X("<button class=\"svelte-1td2jxo\">Auto all</button>"), Ai = /* @__PURE__ */ X("<tr class=\"grp click svelte-1td2jxo\"><td class=\"mono c-id svelte-1td2jxo\"><span class=\"caret svelte-1td2jxo\"> </span> <span class=\"faint gap svelte-1td2jxo\"> </span> <!> <!> <!> <!> <!></td><td data-label=\"desired\" class=\"svelte-1td2jxo\"> </td><td data-label=\"open\" class=\"svelte-1td2jxo\"> </td><td data-label=\"working\" class=\"svelte-1td2jxo\"> </td><td data-label=\"idle\" class=\"svelte-1td2jxo\"> </td><td class=\"svelte-1td2jxo\"></td><td class=\"svelte-1td2jxo\"></td><td class=\"svelte-1td2jxo\"></td><td class=\"c-act acts svelte-1td2jxo\"><!></td></tr>"), ji = /* @__PURE__ */ X("<tr class=\"click fold svelte-1td2jxo\" data-testid=\"fold\"><td class=\"faint svelte-1td2jxo\" colspan=\"9\"> </td></tr>"), Mi = /* @__PURE__ */ X("<tbody data-testid=\"pool-group\"><!><!><!></tbody>"), Ni = /* @__PURE__ */ X("<tbody><tr><td class=\"faint\" colspan=\"9\"> </td></tr></tbody>"), Pi = /* @__PURE__ */ X("<span class=\"faint gap svelte-1td2jxo\">all waiting</span>"), Fi = /* @__PURE__ */ X("<tr class=\"grp click svelte-1td2jxo\"><td class=\"mono c-id svelte-1td2jxo\" colspan=\"5\"><span class=\"caret svelte-1td2jxo\"> </span> <span class=\"faint gap svelte-1td2jxo\"> </span> <!> <!> <!></td></tr>"), Ii = /* @__PURE__ */ X("<tr class=\"click fold svelte-1td2jxo\"><td class=\"faint svelte-1td2jxo\" colspan=\"5\"> </td></tr>"), Li = /* @__PURE__ */ X("<tbody data-testid=\"trigger-group\"><!><!><!></tbody>"), Ri = /* @__PURE__ */ X("<tbody><tr><td class=\"faint\" colspan=\"5\"> </td></tr></tbody>"), zi = /* @__PURE__ */ X("<div class=\"bar svelte-1td2jxo\"><input type=\"text\" data-testid=\"pool-search\" placeholder=\"search pools and triggers\" aria-label=\"search by name\" class=\"svelte-1td2jxo\"/> <!></div> <!> <!> <h2 class=\"st\">Pools</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"scale\"><thead><tr><th> </th><th>Desired</th><th>Open</th><th>Working</th><th>Idle</th><th>Cordoned</th><th>Min/max</th><th>Target</th><th></th></tr></thead><!></table></div> <h2 class=\"st\">Triggers</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"triggers\"><thead><tr><th>Trigger</th><th>Holds</th><th>Why not / stalls</th><th>Last fired</th><th>Next cron</th></tr></thead><!></table></div>", 1), Bi = /* @__PURE__ */ X("<header class=\"vh\"><h1>Pools and triggers</h1> <!> <div class=\"actions\"><label class=\"faint\" for=\"pool-group\">group by</label> <select id=\"pool-group\" data-testid=\"group-by\"></select></div></header> <div class=\"pad\"><!></div>", 1);
function Vi(e, t) {
	Re(t, !0);
	let n = (e, t = m) => {
		var n = gi(), r = L(n), i = L(r), a = z(i), o = (e) => {
			Z(e, li());
		};
		$(a, (e) => {
			t().quarantined && e(o);
		});
		var s = z(a, 2), c = (e) => {
			Z(e, ui());
		};
		$(s, (e) => {
			t().paused && e(c);
		});
		var l = z(s, 2), u = (e) => {
			Z(e, di());
		};
		$(l, (e) => {
			t().drained && e(u);
		});
		var d = z(l, 2), f = (e) => {
			var n = fi(), r = R(n, !0);
			B(() => Q(r, t().why)), Z(e, n);
		};
		$(d, (e) => {
			t().why && e(f);
		});
		var p = z(d, 2), h = (e) => {
			var n = fi(), r = R(n);
			B((e) => Q(r, `until ${e ?? ""}`), [() => Zr(String(t().until))]), Z(e, n);
		};
		$(p, (e) => {
			t().until && e(h);
		}), A(r);
		var g = z(r), _ = R(g, !0), v = z(g), y = R(v, !0), b = z(v), x = R(b, !0), S = z(b), ee = R(S, !0), C = z(S), ne = R(C, !0), T = z(C), re = R(T), ie = z(T), ae = L(ie), oe = (e) => {
			var n = pi(), r = R(n);
			B(() => Q(r, `target ${t().target ?? ""}`)), Z(e, n);
		}, E = (e) => {
			Z(e, mi());
		};
		$(ae, (e) => {
			t().target == null ? e(E, -1) : e(oe);
		}), A(ie);
		var se = z(ie), ce = L(se), le = z(ce, 2), ue = (e) => {
			var n = hi();
			Zn("click", n, () => te({
				pool: t().name,
				auto: !0
			})), Z(e, n);
		};
		$(le, (e) => {
			t().target != null && e(ue);
		}), A(se), A(n), B(() => {
			Ur(n, "data-pool", t().name), Q(i, `${t().name ?? ""} `), Q(_, t().desired), Q(y, t().open), Q(x, t().working), Q(ee, t().idle), Q(ne, t().cordoned), Q(re, `${t().min ?? ""}/${t().max ?? ""}`);
		}), Zn("click", ce, () => w(t())), Z(e, n);
	}, r = Yr(() => t.api), i = (e) => t.api.go(e), a = Jr(() => t.api, () => "api/pools"), o = /* @__PURE__ */ jt(""), s = /* @__PURE__ */ N(() => r.group || "instance"), c = /* @__PURE__ */ N(() => r.q ?? ""), l = /* @__PURE__ */ N(() => r.label ? r.label.split(",") : []), u = /* @__PURE__ */ N(() => new Set(r.open ? r.open.split(",") : [])), d = /* @__PURE__ */ N(() => a.data ? $r(a.data) : []), f = /* @__PURE__ */ N(() => a.data?.triggers ?? []), p = /* @__PURE__ */ N(() => ri([...Y(d), ...Y(f)])), h = /* @__PURE__ */ N(() => ai([...Y(d), ...Y(f)])), g = /* @__PURE__ */ N(() => ii(Y(d), Y(c), Y(l))), v = /* @__PURE__ */ N(() => ii(Y(f), Y(c), Y(l))), y = /* @__PURE__ */ N(() => oi(Y(g), Y(s))), b = /* @__PURE__ */ N(() => si(Y(v), Y(s))), x = /* @__PURE__ */ N(() => Y(d).filter((e) => !e.desired && !e.open && !e.working && !e.quarantined && !e.paused && !e.drained && e.target == null).length);
	function S(e) {
		let t = new Set(Y(u));
		t.delete(e) || t.add(e), i({ open: [...t].join(",") });
	}
	let ee = (e, t = "") => t + (e || "_");
	function C(e) {
		let t = Y(l).includes(e) ? Y(l).filter((t) => t !== e) : [...Y(l), e];
		i({ label: t.join(",") });
	}
	async function te(e) {
		F(o, "");
		try {
			await t.api.post("api/pool/scale", e), a.refresh();
		} catch (e) {
			F(o, Xr(e), !0);
		}
	}
	function w(e) {
		let t = window.prompt("Target sessions for " + e.name + " (" + e.min + ".." + e.max + ")", String(e.target ?? e.desired));
		if (t === null || t.trim() === "") return;
		let n = Number(t);
		!Number.isInteger(n) || n < 0 ? F(o, "target must be a whole number") : te({
			pool: e.name,
			target: n
		});
	}
	async function ne(e) {
		F(o, "");
		try {
			for (let n of e) n.target != null && await t.api.post("api/pool/scale", {
				pool: n.name,
				auto: !0
			});
			a.refresh();
		} catch (e) {
			F(o, Xr(e), !0);
		}
	}
	var T = Bi(), re = Xt(T), ie = z(L(re), 2), ae = (e) => {
		var t = yi(), n = L(t), r = z(n), i = (e) => {
			var t = or();
			B(() => Q(t, `· host max ${a.data.host_max ?? ""}`)), Z(e, t);
		};
		$(r, (e) => {
			a.data.host_max && e(i);
		}), A(t), B(() => Q(n, `${Y(d).length ?? ""} pools, ${Y(x) ?? ""} idle`)), Z(e, t);
	};
	$(ie, (e) => {
		a.data && e(ae);
	});
	var oe = z(ie, 2), E = z(L(oe), 2);
	Sr(E, 20, () => Y(p), (e) => e, (e, t) => {
		var n = bi(), r = R(n, !0), i = {};
		B(() => {
			Q(r, t), i !== (i = t) && (n.value = (n.__value = i) ?? "");
		}), Z(e, n);
	}), A(E);
	var se;
	Pr(E), A(oe), A(re);
	var ce = z(re, 2), le = L(ce), ue = (e) => {
		var t = xi(), n = R(t, !0);
		B(() => Q(n, a.error)), Z(e, t);
	}, de = (e) => {
		var t = zi(), r = Xt(t), a = L(r);
		Vr(a);
		var p = z(a, 2), m = (e) => {
			var t = Si();
			Zn("click", t, () => i({
				q: "",
				label: ""
			})), Z(e, t);
		};
		$(p, (e) => {
			(Y(c) || Y(l).length) && e(m);
		}), A(r);
		var g = z(r, 2), v = (e) => {
			var t = wi();
			Sr(t, 21, () => Y(h), ([e, t]) => e, (e, t) => {
				var n = /* @__PURE__ */ N(() => _(Y(t), 2));
				let r = () => Y(n)[0], i = () => Y(n)[1];
				var a = Ci();
				let o;
				var s = L(a), c = R(z(s), !0);
				A(a), B((e) => {
					o = Ar(a, 1, "", null, o, { on: e }), Q(s, `${r() ?? ""} `), Q(c, i());
				}, [() => Y(l).includes(r())]), Zn("click", a, () => C(r())), Z(e, a);
			}), A(t), Z(e, t);
		};
		$(g, (e) => {
			Y(h).length && e(v);
		});
		var x = z(g, 2), te = (e) => {
			var t = xi(), n = R(t, !0);
			B(() => Q(n, Y(o))), Z(e, t);
		};
		$(x, (e) => {
			Y(o) && e(te);
		});
		var w = z(x, 4), T = L(w), re = L(T), ie = L(re), ae = R(L(ie), !0);
		xe(8), A(ie), A(re), Sr(z(re), 17, () => Y(y), (e) => e.name, (e, t) => {
			let r = /* @__PURE__ */ N(() => ee(Y(t).name)), i = /* @__PURE__ */ N(() => Y(u).has(Y(r)));
			var a = Mi(), o = L(a), c = (e) => {
				var n = Ai(), a = L(n), o = L(a), s = R(o, !0), c = z(o), l = z(c), u = R(l), d = z(l, 2), f = (e) => {
					var n = Ti(), r = R(n);
					B(() => Q(r, `${Y(t).pools.length ?? ""} idle`)), Z(e, n);
				};
				$(d, (e) => {
					Y(t).quiet && e(f);
				});
				var p = z(d, 2), m = (e) => {
					var n = Ei(), r = R(n);
					B(() => Q(r, `${Y(t).quarantined ?? ""} quarantined`)), Z(e, n);
				};
				$(p, (e) => {
					Y(t).quarantined && e(m);
				});
				var h = z(p, 2), g = (e) => {
					var n = Di(), r = R(n);
					B(() => Q(r, `${Y(t).paused ?? ""} paused`)), Z(e, n);
				};
				$(h, (e) => {
					Y(t).paused && e(g);
				});
				var _ = z(h, 2), v = (e) => {
					var n = Di(), r = R(n);
					B(() => Q(r, `${Y(t).drained ?? ""} drained`)), Z(e, n);
				};
				$(_, (e) => {
					Y(t).drained && e(v);
				});
				var y = z(_, 2), b = (e) => {
					var n = Oi(), r = R(n);
					B(() => Q(r, `${Y(t).targeted ?? ""} targeted`)), Z(e, n);
				};
				$(y, (e) => {
					Y(t).targeted && e(b);
				}), A(a);
				var x = z(a), ee = R(x, !0), C = z(x), te = R(C, !0), w = z(C), T = R(w, !0), re = z(w), ie = R(re, !0), ae = z(re, 4), oe = L(ae), E = (e) => {
					var n = ki();
					Zn("click", n, (e) => {
						e.stopPropagation(), ne(Y(t).pools);
					}), Z(e, n);
				};
				$(oe, (e) => {
					Y(t).targeted && e(E);
				}), A(ae), A(n), B(() => {
					Ur(n, "aria-expanded", Y(i)), Q(s, Y(i) || !Y(t).folded.length ? "▾" : "▸"), Q(c, ` ${Y(t).name ?? ""} `), Q(u, `${Y(t).pools.length ?? ""} pool${Y(t).pools.length === 1 ? "" : "s"}`), Q(ee, Y(t).desired), Q(te, Y(t).open), Q(T, Y(t).working), Q(ie, Y(t).idle);
				}), Zn("click", n, () => S(Y(r))), Z(e, n);
			};
			$(o, (e) => {
				Y(s) !== "none" && e(c);
			});
			var l = z(o);
			Sr(l, 17, () => Y(i) ? Y(t).pools : Y(t).shown, (e) => e.name, (e, t) => {
				n(e, () => Y(t));
			});
			var d = z(l), f = (e) => {
				var n = ji(), i = R(L(n));
				A(n), B(() => Q(i, `${Y(t).folded.length ?? ""} idle pool${Y(t).folded.length === 1 ? "" : "s"}`)), Zn("click", n, () => S(Y(r))), Z(e, n);
			};
			$(d, (e) => {
				Y(i) || !Y(t).folded.length || Y(t).quiet && Y(s) !== "none" || e(f);
			}), A(a), B(() => Ur(a, "data-group", Y(t).name)), Z(e, a);
		}, (e) => {
			var t = Ni(), n = L(t), r = R(L(n), !0);
			A(n), A(t), B(() => Q(r, Y(d).length ? "No pools match." : "No ticket pools.")), Z(e, t);
		}), A(T), A(w);
		var oe = z(w, 4), E = L(oe);
		Sr(z(L(E)), 17, () => Y(b), (e) => e.name, (e, t) => {
			let n = /* @__PURE__ */ N(() => ee(Y(t).name, "t:")), r = /* @__PURE__ */ N(() => Y(u).has(Y(n)));
			var i = Li(), a = L(i), o = (e) => {
				var i = Fi(), a = L(i), o = L(a), s = R(o, !0), c = z(o), l = z(c), u = R(l), d = z(l, 2), f = (e) => {
					var n = Ei(), r = R(n);
					B(() => Q(r, `${Y(t).stalled ?? ""} stalled`)), Z(e, n);
				};
				$(d, (e) => {
					Y(t).stalled && e(f);
				});
				var p = z(d, 2), m = (e) => {
					var n = Oi(), r = R(n);
					B(() => Q(r, `${Y(t).firing ?? ""} firing`)), Z(e, n);
				};
				$(p, (e) => {
					Y(t).firing && e(m);
				});
				var h = z(p, 2), g = (e) => {
					Z(e, Pi());
				};
				$(h, (e) => {
					Y(t).quiet && e(g);
				}), A(a), A(i), B(() => {
					Ur(i, "aria-expanded", Y(r)), Q(s, Y(r) || !Y(t).folded.length ? "▾" : "▸"), Q(c, ` ${Y(t).name ?? ""} `), Q(u, `${Y(t).triggers.length ?? ""} trigger${Y(t).triggers.length === 1 ? "" : "s"}`);
				}), Zn("click", i, () => S(Y(n))), Z(e, i);
			};
			$(a, (e) => {
				Y(s) !== "none" && e(o);
			});
			var c = z(a);
			Sr(c, 17, () => Y(r) ? Y(t).triggers : Y(t).shown, (e) => e.name, (e, t) => {
				ci(e, () => Y(t));
			});
			var l = z(c), d = (e) => {
				var r = Ii(), i = R(L(r));
				A(r), B(() => Q(i, `${Y(t).folded.length ?? ""} waiting`)), Zn("click", r, () => S(Y(n))), Z(e, r);
			};
			$(l, (e) => {
				Y(r) || !Y(t).folded.length || Y(t).quiet && Y(s) !== "none" || e(d);
			}), A(i), B(() => Ur(i, "data-group", Y(t).name)), Z(e, i);
		}, (e) => {
			var t = Ri(), n = L(t), r = R(L(n), !0);
			A(n), A(t), B(() => Q(r, Y(f).length ? "No triggers match." : "No triggers.")), Z(e, t);
		}), A(E), A(oe), B(() => {
			Hr(a, Y(c)), Q(ae, Y(s) === "none" ? "Pool" : Y(s) + " / pool");
		}), Zn("change", a, (e) => i({ q: e.currentTarget.value.trim() })), Z(e, t);
	};
	$(le, (e) => {
		a.error ? e(ue) : a.data && e(de, 1);
	}), A(ce), B(() => {
		se !== (se = Y(s)) && (E.value = (E.__value = se) ?? "", Nr(E, se));
	}), Zn("change", E, (e) => i({
		group: e.currentTarget.value === "instance" ? "" : e.currentTarget.value,
		open: ""
	})), Z(e, T), ze();
}
// plugins/pools/index.js
Qn(["click", "change"]), qr("desk-pools-view", Vi);
