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
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, C = 65536, te = 1 << 19, ne = 1 << 20, re = 1 << 25, ie = 1 << 21, ae = 1 << 22, oe = 1 << 23, se = Symbol("$state"), ce = Symbol("component"), le = Symbol("legacy props"), ue = Symbol(""), de = Symbol("attributes"), fe = Symbol("class"), pe = Symbol("style"), me = Symbol("text"), he = new class extends Error {
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
var w = !1;
function be(e) {
	w = e;
}
var T;
function E(t) {
	if (t === null) throw ve(), e;
	return T = t;
}
function xe() {
	return E(/* @__PURE__ */ Wt(T));
}
function D(t) {
	if (w) {
		if (/* @__PURE__ */ Wt(T) !== null) throw ve(), e;
		T = t;
	}
}
function Se(e = 1) {
	if (w) {
		for (var t = e, n = T; t--;) n = /* @__PURE__ */ Wt(n);
		T = n;
	}
}
function Ce(e = !0) {
	for (var t = 0, n = T;;) {
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
// node_modules/svelte/src/internal/client/context.js
var O = null;
function ze(e) {
	O = e;
}
function Be(e, t = !1, n) {
	O = {
		p: O,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: G,
		l: null
	};
}
function Ve(e) {
	var t = O, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) an(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, O = t.p, He(e);
}
function He(e = {}) {
	return s(e, ce, { value: !0 }), e;
}
function Ue() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var We = [];
function Ge() {
	var e = We;
	We = [], h(e);
}
function Ke(e) {
	if (We.length === 0 && !mt) {
		var t = We;
		queueMicrotask(() => {
			t === We && Ge();
		});
	}
	We.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var qe = ~(v | y | _);
function k(e, t) {
	e.f = e.f & qe | t;
}
function Je(e) {
	e.f & 512 || e.deps === null ? k(e, _) : k(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ye(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), k(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Xe(e) {
	var t = H, n = G;
	W(null), Cn(null);
	try {
		return e();
	} finally {
		W(t), Cn(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ze(e, t, n, r) {
	let i = Ue() ? tt : it;
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
				Zt(e, s);
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
		Promise.all(n.map((e) => /* @__PURE__ */ rt(e))).then(u).catch((e) => Zt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), $e());
	}) : f();
}
function Qe() {
	var e = G, t = H, n = O, r = j;
	return function(i = !0) {
		Cn(e), W(t), ze(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function $e(e = !0) {
	Cn(null), W(null), ze(null), e && j?.deactivate();
}
function et() {
	var e = G, t = e.b, n = j, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function tt(e) {
	var n = 2 | v;
	return G !== null && (G.f |= te), {
		ctx: O,
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
var nt = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function rt(e, n, r) {
	let i = G;
	i === null && Oe();
	var a = void 0, o = At(t), s = !H, c = /* @__PURE__ */ new Set();
	return sn(() => {
		var t = G, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== he && n.reject(e);
			}).finally($e);
		} catch (e) {
			n.reject(e), $e();
		}
		var r = j;
		if (s) {
			if (t.f & 32768) var l = et();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(nt);
			else for (let e of c.values()) e.reject(nt);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== nt && (r.activate(), t ? (o.f |= oe, Pt(o, t)) : (o.f & 8388608 && (o.f ^= oe), Pt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), nn(() => {
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
function A(e) {
	let t = /* @__PURE__ */ tt(e);
	return Tn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function it(e) {
	let t = /* @__PURE__ */ tt(e);
	return t.equals = De, t;
}
function at(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) V(t[n]);
	}
}
function ot(e) {
	var n, r = G, i = e.parent;
	if (!xn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	Cn(i);
	try {
		at(e), n = Pn(e);
	} finally {
		Cn(r);
	}
	return n;
}
function st(e) {
	var t = ot(e);
	!e.equals(t) && (e.wv = jn(), (!j?.is_fork || e.deps === null) && (j === null ? e.v = t : (j.capture(e, t, !0), dt?.capture(e, t, !0)), e.deps === null)) ? k(e, _) : xn || (ft === null ? Je(e) : (tn() || j?.is_fork) && ft.set(e, t));
}
function ct(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Xe(() => {
		t.ac.abort(he), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Ln(t, 0), dn(t));
}
function lt(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Rn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ut = null, j = null, dt = null, ft = null, pt = null, mt = !1, ht = !1, gt = null, _t = null, vt = 0, yt = 1, bt = class e {
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
		for (var t = gt = [], n = [], r = _t = []; this.#a.length > 0;) {
			vt++ > 1e3 && (this.#b(), xt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Et(e), this.#d() || this.discard(), t;
			}
		}
		if (j = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (gt = null, _t = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Tt(e, t);
			r.length > 0 && j.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), dt = this, Ct(n), Ct(t), dt = null, this.#i?.resolve();
			var o = j;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ot.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : Mn(r) && (i & 16 && this.add_dirty_reaction(r, y), Rn(r));
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
		this.oncommit(() => e.discard()), e.#b(), j = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Ye(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), ft?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		j = this;
	}
	deactivate() {
		j = null, ft = null;
	}
	flush() {
		try {
			ht = !0, j = this, this.#m();
		} finally {
			vt = 0, pt = null, gt = null, _t = null, ht = !1, j = null, ft = null, Ot.clear();
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
		this.#u || (this.#u = !0, Ke(() => {
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
		if (j === null) {
			let t = j = new e();
			!ht && Ke(() => {
				t.started || t.flush();
			});
		}
		return j;
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
		Ne();
	} catch (e) {
		Zt(e, pt);
	}
}
var St = null;
function Ct(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Mn(r) && (St = /* @__PURE__ */ new Set(), Rn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && mn(r), St?.size > 0)) {
				Ot.clear();
				for (let e of St) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) St.has(n) && (St.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Rn(n);
					}
				}
				St.clear();
			}
		}
		St = null;
	}
}
function wt(e) {
	j.schedule(e);
}
function Tt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), k(e, _);
		for (var n = e.first; n !== null;) Tt(n, t), n = n.next;
	}
}
function Et(e) {
	k(e, _);
	for (var t = e.first; t !== null;) Et(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Dt = /* @__PURE__ */ new Set(), Ot = /* @__PURE__ */ new Map(), kt = !1;
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
function M(e, t) {
	let n = At(e, t);
	return Tn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function jt(e, t = !1, n = !0) {
	let r = At(e);
	return t || (r.equals = De), r;
}
function N(e, t, n = !1) {
	return H !== null && (!U || H.f & 131072) && Ue() && H.f & 4325394 && (wn === null || !wn.has(e)) && Le(), Pt(e, n ? P(t) : t, _t);
}
var Mt = null, Nt = 0;
function Pt(e, t, n = null) {
	if (!e.equals(t)) {
		xn ? Ot.set(e, t) : Ot.has(e) || Ot.set(e, e.v);
		var r = bt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ot(t), ft === null && Je(t);
		}
		e.wv = jn(), Mt = null, Nt = 0, Lt(e, v, n), Mt = null, Ue() && G !== null && G.f & 1024 && !(G.f & 96) && (J === null ? En([e]) : J.push(e)), !r.is_fork && Dt.size > 0 && !kt && Ft();
	}
	return t;
}
function Ft() {
	kt = !1;
	for (let e of Dt) {
		e.f & 1024 && k(e, y);
		let t;
		try {
			t = Mn(e);
		} catch {
			t = !0;
		}
		t && Rn(e);
	}
	Dt.clear();
}
function It(e) {
	N(e, e.v + 1);
}
function Lt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ue(), a = r.length;
		if (Nt += a, Nt > 1e5 && Mt === null && (Mt = /* @__PURE__ */ new Set()), Mt !== null) {
			if (Mt.has(e)) return;
			Mt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== G) {
				var l = (c & v) === 0;
				if (l && k(s, t), c & 131072) Dt.add(s);
				else if (c & 2) {
					var u = s;
					ft?.delete(u), Lt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && St !== null && St.add(d), n === null ? wt(d) : n.push(d);
				}
			}
		}
	}
}
function P(e) {
	if (typeof e != "object" || !e || se in e || ce in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ M(0), s = null, l = kn, p = (e) => {
		if (kn === l) return e();
		var t = H, n = kn;
		W(null), An(l);
		var r = e();
		return W(t), An(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ M(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Fe();
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
					i.set(n, e), It(o);
				}
			} else N(r, t), It(o);
			return !0;
		},
		get(n, r, a) {
			if (r === se) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ M(P(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
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
			if (n === se) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (G !== null || H !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ M(a ? P(e[n]) : t, s)), i.set(n, r)), Y(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ M(t, s)), i.set(f + "", m)) : N(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ M(void 0, s)), N(u, P(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => P(r));
				N(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && N(_, v + 1);
				}
				It(o);
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
			Ie();
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
function F(e = "") {
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
function I(e, t) {
	if (!w) return /* @__PURE__ */ Ut(e);
	var n = /* @__PURE__ */ Ut(T);
	if (n === null) n = T.appendChild(F());
	else if (t && n.nodeType !== 3) {
		var r = F();
		return n?.before(r), E(r), r;
	}
	return t && Yt(n), E(n), n;
}
function Gt(e, t = !1) {
	if (!w) {
		var n = /* @__PURE__ */ Ut(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Wt(n) : n;
	}
	if (t) {
		if (T?.nodeType !== 3) {
			var r = F();
			return T?.before(r), E(r), r;
		}
		Yt(T);
	}
	return T;
}
function L(e, t = !1) {
	if (!w) return /* @__PURE__ */ Ut(e);
	var n = I(e, t);
	return D(e), n;
}
function R(e, t = 1, n = !1) {
	let r = w ? T : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Wt(r);
	if (!w) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = F();
			return r === null ? i?.after(a) : r.before(a), E(a), a;
		}
		Yt(r);
	}
	return E(r), r;
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
	var t = G;
	if (t === null) return H.f |= oe, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	Zt(e, t);
}
function Zt(t, n) {
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
function Qt(e) {
	G === null && (H === null && Me(e), je()), xn && Ae(e);
}
function $t(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function en(e, t) {
	var n = G;
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
	j?.register_created_effect(r);
	var i = r;
	if (e & 4) gt === null ? bt.ensure().schedule(r) : gt.push(r);
	else if (t !== null) {
		try {
			Rn(r);
		} catch (e) {
			throw V(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && $t(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function tn() {
	return H !== null && !U;
}
function nn(e) {
	let t = en(8, null);
	return k(t, _), t.teardown = e, t;
}
function rn(e) {
	Qt("$effect");
	var t = G.f;
	if (!H && t & 32 && O !== null && !O.i) {
		var n = O;
		(n.e ??= []).push(e);
	} else return an(e);
}
function an(e) {
	return en(4 | ne, e);
}
function on(e) {
	bt.ensure();
	let t = en(64 | te, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? hn(t, () => {
			V(t), n(void 0);
		}) : (V(t), n(void 0));
	});
}
function sn(e) {
	return en(ae | te, e);
}
function cn(e, t = 0) {
	return en(8 | t, e);
}
function z(e, t = [], n = [], r = []) {
	Ze(r, t, n, (t) => {
		en(8, () => {
			e(...t.map(Y));
		});
	});
}
function ln(e, t = 0) {
	return en(16 | t, e);
}
function B(e) {
	return en(32 | te, e);
}
function un(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = xn, r = H;
		Sn(!0), W(null);
		try {
			t.call(null);
		} catch (t) {
			Zt(t, e.parent);
		} finally {
			Sn(n), W(r);
		}
	}
}
function dn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Xe(() => {
			e.abort(he);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : V(n, t), n = r;
	}
}
function fn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || V(t), t = n;
	}
}
function V(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (pn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, dn(e, t && !n), Ln(e, 0);
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
		n && V(e), t && t();
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
		e.f ^= b, e.f & 1024 || (k(e, v), bt.ensure().schedule(e));
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
var H = null, U = !1;
function W(e) {
	H = e;
}
var G = null;
function Cn(e) {
	G = e;
}
var wn = null;
function Tn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (wn ??= /* @__PURE__ */ new Set()).add(e);
}
var K = null, q = 0, J = null;
function En(e) {
	J = e;
}
var Dn = 1, On = 0, kn = On;
function An(e) {
	kn = e;
}
function jn() {
	return ++Dn;
}
function Mn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Mn(a) && st(a), a.wv > e.wv) return !0;
		}
		t & 512 && ft === null && k(e, _);
	}
	return !1;
}
function Nn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(wn !== null && wn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Nn(a, t, !1) : t === a && (n ? k(a, v) : a.f & 1024 && k(a, y), wt(a));
	}
}
function Pn(e) {
	var t = K, n = q, r = J, i = H, a = wn, o = O, s = U, c = kn, l = e.f;
	K = null, q = 0, J = null, H = l & 96 ? null : e, wn = null, ze(e.ctx), U = !1, kn = ++On, e.ac !== null && (Xe(() => {
		e.ac.abort(he);
	}), e.ac = null);
	try {
		e.f |= ie;
		var u = e.fn, d = u();
		e.f |= S;
		var f = Fn(e);
		if (Ue() && J !== null && !U && f !== null && !(e.f & 6146)) for (var p = 0; p < J.length; p++) Nn(J[p], e);
		if (i !== null && i !== e) {
			if (On++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = On;
			if (t !== null) for (let e of t) e.rv = On;
			J !== null && (r === null ? r = J : r.push(...J));
		}
		return e.f & 8388608 && (e.f ^= oe), d;
	} catch (t) {
		return Fn(e), Xt(t);
	} finally {
		e.f ^= ie, K = t, q = n, J = r, H = i, wn = a, ze(o), U = s, kn = c;
	}
}
function Fn(e) {
	var t = e.deps, n = j?.is_fork;
	if (K !== null) {
		var r;
		if (n || Ln(e, q), t !== null && q > 0) for (t.length = q + K.length, r = 0; r < K.length; r++) t[q + r] = K[r];
		else e.deps = t = K;
		if (tn() && e.f & 512) for (r = q; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && q < t.length && (Ln(e, q), t.length = q);
	return t;
}
function In(e, n) {
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
		c.f & 512 && (c.f ^= 512), c.v !== t && Je(c), c.ac !== null && Xe(() => {
			c.ac.abort(he), c.ac = null, k(c, v);
		}), ct(c), Ln(c, 0);
	}
}
function Ln(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) In(e, n[r]);
}
function Rn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		k(e, _);
		var n = G;
		G = e;
		try {
			t & 16777232 ? fn(e) : dn(e), un(e);
			var r = Pn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Dn;
		} finally {
			G = n;
		}
	}
}
function Y(e) {
	var t = !!(e.f & 2);
	if (bn?.add(e), H !== null && !U && !(G !== null && G.f & 16384) && (wn === null || !wn.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < On && (e.rv = On, K === null && n !== null && n[q] === e ? q++ : K === null ? K = [e] : K.push(e));
		else {
			H.deps ??= [], a.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : a.call(r, H) || r.push(H);
		}
	}
	if (xn && Ot.has(e)) return Ot.get(e);
	if (t) {
		var i = e;
		if (xn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Bn(i)) && (o = ot(i)), Ot.set(i, o), o;
		}
		var s = !(i.f & 512) && !U && H !== null && !!(H.f & 512), c = (i.f & S) === 0;
		Mn(i) && (s && (i.f |= 512), st(i)), s && !c && (lt(i), zn(i));
	}
	if (ft?.has(e)) return ft.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function zn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (lt(n), zn(n));
	}
}
function Bn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ot.has(t) || t.f & 2 && Bn(t)) return !0;
	return !1;
}
function Vn(e) {
	var t = U;
	try {
		return U = !0, e();
	} finally {
		U = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Hn = Symbol("events"), Un = /* @__PURE__ */ new Set(), Wn = /* @__PURE__ */ new Set();
function Gn(e, t, n) {
	(t[Hn] ??= {})[e] = n;
}
function Kn(e) {
	for (var t = 0; t < e.length; t++) Un.add(e[t]);
	for (var n of Wn) n(e);
}
var qn = null, Jn = !1;
function Yn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	qn = e, Jn || (Jn = !0, setTimeout(() => {
		Jn = !1, qn = null;
	}));
	var o = 0, c = qn === e && e[Hn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Hn] = t;
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
		W(null), Cn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Hn]?.[r];
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
			e[Hn] = t, delete e.currentTarget, W(d), Cn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Xn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Zn(e) {
	return Xn?.createHTML(e) ?? e;
}
function Qn(e) {
	var t = Jt("template");
	return t.innerHTML = Zn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function $n(e, t) {
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
		if (w) return $n(T, null), T;
		i === void 0 && (i = Qn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Ut(i)));
		var t = r || zt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Ut(t), s = t.lastChild;
			$n(o, s);
		} else $n(t, t);
		return t;
	};
}
function er(e = "") {
	if (!w) {
		var t = F(e + "");
		return $n(t, t), t;
	}
	var n = T;
	return n.nodeType === 3 ? Yt(n) : (n.before(n = F()), E(n)), $n(n, n), n;
}
function tr() {
	if (w) return $n(T, null), T;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = F();
	return e.append(t, n), $n(t, n), e;
}
function Z(e, t) {
	if (w) {
		var n = G;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = T), xe();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var nr = ["touchstart", "touchmove"];
function rr(e) {
	return nr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function ir(e) {
	let t = 0, n = At(0), r;
	return () => {
		tn() && (Y(n), cn(() => (t === 0 && (r = Vn(() => e(() => It(n)))), t += 1, () => {
			Ke(() => {
				--t, t === 0 && (r?.(), r = void 0, It(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var ar = C | te;
function or(e, t, n, r) {
	new sr(e, t, n, r);
}
var sr = class {
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
	#m = ir(() => (this.#p = At(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = G;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = G.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = ln(() => {
			if (w) {
				let e = we(this.#t);
				xe();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, ar), w && (this.#e = T);
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
		Ke(r), t && (this.#s = B(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Re(), this.#s !== null && hn(this.#s, () => {
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
					Zt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = B(() => e(this.#e)), Ke(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = F(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return B(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							Zt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(j)) : this.#u === 0 && (this.#e.before(e), this.#c = null, hn(this.#o, () => {
					this.#o = null;
				}), this.#x(j));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = B(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				yn(this.#a, e);
				let t = this.#n.pending;
				this.#o = B(() => t(this.#e));
			} else this.#x(j);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Ye(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = G, n = H, r = O;
		Cn(this.#i), W(this.#i), ze(this.#i.ctx);
		try {
			return bt.ensure(), e();
		} finally {
			Cn(t), W(n), ze(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && hn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ke(() => {
			this.#d = !1, this.#p && Pt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Y(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		j?.is_fork ? (this.#a && j.skip_effect(this.#a), this.#o && j.skip_effect(this.#o), this.#s && j.skip_effect(this.#s), j.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (V(this.#a), null), this.#o &&= (V(this.#o), null), this.#s &&= (V(this.#s), null), w && (E(this.#t), Se(), E(Ce()));
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
					return Zt(e, this.#i.parent), null;
				}
			}));
		};
		Ke(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					Zt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => Zt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Q(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[me] ??= e.nodeValue) && (e[me] = n, e.nodeValue = `${n}`);
}
function cr(e, t) {
	return ur(e, t);
}
var lr = /* @__PURE__ */ new Map();
function ur(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Ht();
	var u = void 0, d = on(() => {
		var c = r ?? n.appendChild(F());
		or(c, { pending: () => {} }, (n) => {
			Be({});
			var r = O;
			if (s && (r.c = s), a && (i.$$events = a), w && $n(n, null), u = t(n, i) || He(), w && (G.nodes.end = T, T === null || T.nodeType !== 8 || T.data !== "]")) throw ve(), e;
			Ve();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = rr(r);
					for (let e of [n, document]) {
						var a = lr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), lr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Yn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Un)), Wn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = lr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Yn), t.delete(e), t.size === 0 && lr.delete(r)) : t.set(e, i);
			}
			Wn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return dr.set(u, d), u;
}
var dr = /* @__PURE__ */ new WeakMap();
function fr(e, t) {
	let n = dr.get(e);
	return n ? (dr.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var pr = class {
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
				r && (V(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						yn(r, t), t.append(F()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else V(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), hn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (V(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = j, r = qt();
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
		} else w && (this.anchor = T), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function $(e, t, n = !1) {
	var r;
	w && (r = T, xe());
	var i = new pr(e), a = n ? C : 0;
	function o(e, t) {
		if (w) {
			var n = we(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Ce();
				E(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
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
function mr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		hn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					hr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
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
		hr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function hr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= re, yn(a, document.createDocumentFragment())) : V(t[i], n);
	}
}
var gr;
function _r(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = w ? E(/* @__PURE__ */ Ut(u)) : u.appendChild(F());
	}
	w && xe();
	var d = null, f = /* @__PURE__ */ it(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Y(f);
			_.fallback = d, yr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= re, xr(d, null, c)) : _n(d) : hn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: ln(() => {
			var e = Y(f), r = e.length;
			let o = !1;
			w && we(c) === "[!" != (r === 0) && (c = Ce(), E(c), be(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = j, v = qt(), y = 0; y < r; y += 1) {
				w && T.nodeType === 8 && T.data === "]" && (c = T, o = !0, be(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Pt(S.v, b), S.i && Pt(S.i, y), v && _.unskip_effect(S.e)) : (S = br(l, m ? c : gr ??= F(), b, x, y, a, t, n), m || (S.e.f |= re), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = B(() => s(c)) : (d = B(() => s(gr ??= F())), d.f |= re)), r > u.size && ke("", "", ""), w && r > 0 && E(Ce()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && be(!0), Y(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, w && (c = T);
}
function vr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function yr(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = vr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (_n(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= re, _ === l) xr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), Sr(e, d, _), Sr(e, _, y), xr(_, y, n), d = _, p = [], m = [], l = vr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) xr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					Sr(e, S.prev, ee.next), Sr(e, d, S), Sr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), xr(_, l, n), Sr(e, _.prev, _.next), Sr(e, _, d === null ? e.effect.first : d.next), Sr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = vr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = vr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (hr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = vr(l.next);
		var te = C.length;
		if (te > 0) {
			var ne = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.fix();
			}
			mr(e, C, ne);
		}
	}
	a && Ke(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function br(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? At(n) : /* @__PURE__ */ jt(n, !1, !1) : null, l = o & 2 ? At(i) : null;
	return {
		v: c,
		i: l,
		e: B(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function xr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Wt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function Sr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var Cr = [..." 	\n\r\f\xA0\v﻿"];
function wr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || Cr.includes(r[o - 1])) && (s === r.length || Cr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Tr(e, t, n, r, i, a) {
	var o = e[fe];
	if (w || o !== n || o === void 0) {
		var s = wr(n, r, a);
		w && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[fe] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Er = Symbol("is custom element"), Dr = Symbol("is html"), Or = ge ? "link" : "LINK";
function kr(e, t, n, r) {
	var i = Ar(e);
	if (!(w && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Or))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[ue] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Mr(e).has(t) ? e[t] = n : (!w || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Ar(e) {
	return e[de] ??= {
		[Er]: e.nodeName.includes("-"),
		[Dr]: e.namespaceURI === n
	};
}
var jr = /* @__PURE__ */ new Map();
function Mr(e) {
	var t = e.getAttribute("is") || e.nodeName, n = jr.get(t);
	if (n) return n;
	jr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/reactivity/store.js
var Nr = !1;
function Pr(e) {
	var t = Nr;
	try {
		return Nr = !1, [e(), Nr];
	} finally {
		Nr = t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/props.js
function Fr(e, t, n, r) {
	var i = !0, a = !!(n & 8), o = !!(n & 16), s = r, l = !0, u = void 0, d = () => o && i ? (u ??= /* @__PURE__ */ tt(r), Y(u)) : (l && (l = !1, s = o ? Vn(r) : r), s);
	let f;
	if (a) {
		var p = se in e || le in e;
		f = c(e, t)?.set ?? (p && t in e ? (n) => e[t] = n : void 0);
	}
	var m, h = !1;
	a ? [m, h] = Pr(() => e[t]) : m = e[t], m === void 0 && r !== void 0 && (m = d(), f && (i && Pe(t), f(m)));
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
	var v = !1, y = (n & 1 ? tt : it)(() => (v = !1, g()));
	a && Y(y);
	var b = G;
	return (function(e, t) {
		if (arguments.length > 0) {
			let n = t ? Y(y) : i && a ? P(e) : e;
			return N(y, n), v = !0, s !== void 0 && (s = n), e;
		}
		return xn && v || b.f & 16384 ? y.v : Y(y);
	});
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/attention.js
function Ir(e, t, n) {
	if (!t) return [];
	let r = (t, n) => () => e.call(t, n), i = () => e.post("api/attention/ack", {
		kind: t,
		id: n
	});
	switch (t) {
		case "run": return [{
			label: "Resume run",
			run: r("run.resume", { id: n }),
			ask: "Resume run " + n + "?"
		}];
		case "pool": return [{
			label: "Unquarantine pool",
			run: r("pool.unquarantine", { pool: n }),
			ask: "Unquarantine pool " + n + "?"
		}];
		case "wait": return [{
			label: "Clear wait",
			run: r("wait.clear", {
				id: n,
				reason: "cleared in Desk"
			}),
			ask: "Clear the wait on " + n + "?"
		}];
		case "decision": return [
			{
				label: "Approve as target",
				run: r("decision.resolve", {
					id: n,
					action: "approve"
				}),
				ask: "Approve " + n + ": relabel the issue proposed to target?"
			},
			{
				label: "Reject",
				run: r("decision.resolve", {
					id: n,
					action: "reject"
				}),
				ask: "Reject " + n + ": close the issue as not planned?"
			},
			{
				label: "Dismiss",
				run: r("decision.resolve", {
					id: n,
					action: "dismiss"
				}),
				ask: "Dismiss " + n + " for now? The issue is left as is."
			}
		];
		case "ticket":
		case "question":
		case "stale": return [{
			label: "Close ticket",
			run: r("ticket.close", { id: n }),
			ask: "Close " + n + " as done?"
		}];
		default: return [{
			label: "Acknowledge",
			run: i,
			ask: "Acknowledge " + t + ":" + n + "? It is hidden until it recurs."
		}];
	}
}
async function Lr(e, t) {
	if (!await e.confirm(t.ask)) return !1;
	try {
		return await t.run(), e.toast(t.label + ": done"), !0;
	} catch (n) {
		return e.toast(t.label + " failed: " + String(n.message || n)), !1;
	}
}
var Rr = Symbol.for("jobshop.desk.attention.pending"), zr = "jobshop-attention-pending";
function Br() {
	let e = globalThis;
	return e[Rr] ??= /* @__PURE__ */ new Map();
}
function Vr(e, t) {
	Br().set(e, t), globalThis.dispatchEvent?.(new CustomEvent(zr, { detail: {
		key: e,
		label: t
	} }));
}
function Hr(e) {
	Br().delete(e);
}
function Ur(e) {
	let t = (t) => e(t.detail.key, t.detail.label);
	return globalThis.addEventListener?.(zr, t), () => globalThis.removeEventListener?.(zr, t);
}
// kit/format.js
function Wr(e) {
	try {
		let t = new URL(e);
		return t.protocol === "https:" || t.protocol === "http:";
	} catch {
		return !1;
	}
}
// kit/Holder.svelte
var Gr = /* @__PURE__ */ X("<span class=\"svelte-bwb1ob\"> </span>"), Kr = /* @__PURE__ */ X("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-pool\"> </span>"), qr = /* @__PURE__ */ X("<span class=\"name svelte-bwb1ob\" data-testid=\"holder-name\"> </span>"), Jr = /* @__PURE__ */ X("<span class=\"faint model svelte-bwb1ob\"> </span>"), Yr = /* @__PURE__ */ X("<span class=\"tool muted svelte-bwb1ob\" data-testid=\"holder-tool\"> </span>"), Xr = /* @__PURE__ */ X("<button type=\"button\" class=\"holder svelte-bwb1ob\" data-testid=\"holder\"><span></span> <!> <!> <!> <!></button>");
function Zr(e, t) {
	Be(t, !0);
	let n = Fr(t, "label", 3, ""), r = /* @__PURE__ */ A(() => t.h.display_name || t.h.id), i = /* @__PURE__ */ A(() => [
		Y(r),
		[t.h.pool, t.h.model].filter(Boolean).join(" · "),
		t.h.busy ? t.h.tool || "working" : "",
		t.h.id
	].filter(Boolean).join("\n") + "\nClick to open read-only");
	var a = Xr(), o = I(a);
	let s;
	var c = R(o, 2), l = (e) => {
		var t = Gr(), r = L(t, !0);
		z(() => Q(r, n())), Z(e, t);
	};
	$(c, (e) => {
		n() && e(l);
	});
	var u = R(c, 2), d = (e) => {
		var n = Kr(), r = L(n, !0);
		z(() => Q(r, t.h.pool)), Z(e, n);
	}, f = (e) => {
		var t = qr(), n = L(t, !0);
		z(() => Q(n, Y(r))), Z(e, t);
	};
	$(u, (e) => {
		t.h.pool ? e(d) : e(f, -1);
	});
	var p = R(u, 2), m = (e) => {
		var n = Jr(), r = L(n, !0);
		z(() => Q(r, t.h.model)), Z(e, n);
	};
	$(p, (e) => {
		t.h.model && e(m);
	});
	var h = R(p, 2), g = (e) => {
		var n = Yr(), r = L(n, !0);
		z(() => Q(r, t.h.tool || "working")), Z(e, n);
	};
	$(h, (e) => {
		t.h.busy && e(g);
	}), D(a), z(() => {
		kr(a, "title", Y(i)), kr(a, "data-session", t.h.id), kr(a, "data-name", Y(r)), s = Tr(o, 1, "dot svelte-bwb1ob", null, s, {
			busy: t.h.busy,
			dead: t.h.alive === !1
		});
	}), Gn("click", a, (e) => {
		e.stopPropagation(), t.onopen(t.h.id);
	}), Z(e, a), Ve();
}
Kn(["click"]);
// plugins/attention/Attention.svelte
var Qr = /* @__PURE__ */ X("<button class=\"ghost\" data-testid=\"push-test\">Test</button>"), $r = /* @__PURE__ */ X("<!> <button data-testid=\"push-toggle\"> </button>", 1), ei = /* @__PURE__ */ X("<span class=\"faint pushna svelte-t4vask\" data-testid=\"push-state\"> </span>"), ti = /* @__PURE__ */ X("<div class=\"card err\"> </div>"), ni = /* @__PURE__ */ X("<p class=\"faint\">Loading…</p>"), ri = /* @__PURE__ */ X("<div class=\"card empty\" data-testid=\"attention-empty\"><div class=\"big\">Nothing needs you.</div> <div class=\"muted\">No escalations, questions, decisions or other items waiting on you.</div></div>"), ii = /* @__PURE__ */ X("<a target=\"_blank\" rel=\"noopener noreferrer\" data-testid=\"attention-ref\"> </a>"), ai = /* @__PURE__ */ X("· <!>", 1), oi = /* @__PURE__ */ X("<div class=\"summary muted svelte-t4vask\" data-testid=\"attention-summary\"> </div>"), si = /* @__PURE__ */ X("<span class=\"muted\" data-testid=\"row-pending\"> </span>"), ci = /* @__PURE__ */ X("<span class=\"warn-text svelte-t4vask\" data-testid=\"row-stuck\"> </span>"), li = /* @__PURE__ */ X("<button type=\"button\" data-testid=\"row-resolve\" class=\"svelte-t4vask\"> </button>"), ui = /* @__PURE__ */ X("<!> <!> <button type=\"button\" class=\"ghost svelte-t4vask\" data-testid=\"open-chat\">Open chat</button>", 1), di = /* @__PURE__ */ X("<tr data-testid=\"attention-row\"><td class=\"c-id svelte-t4vask\"><span class=\"pill warn\"> </span></td><td class=\"c-title svelte-t4vask\"> <div class=\"faint mono\"> <!></div> <!> <div class=\"acts svelte-t4vask\" role=\"presentation\"><!></div></td><td class=\"muted full svelte-t4vask\"> </td><td class=\"muted c-time svelte-t4vask\"> </td></tr>"), fi = /* @__PURE__ */ X("<div class=\"card scroll\"><table class=\"stack\" data-testid=\"attention-list\"><thead><tr><th>Kind</th><th>Item</th><th>Why</th><th>Since</th></tr></thead><tbody></tbody></table></div>"), pi = /* @__PURE__ */ X("<span class=\"step muted svelte-t4vask\"> </span> <!>", 1), mi = /* @__PURE__ */ X("<span class=\"faint\" data-testid=\"no-holder\">waiting for a session</span>"), hi = /* @__PURE__ */ X("<a class=\"mono svelte-t4vask\"> </a>"), gi = /* @__PURE__ */ X("<span class=\"note svelte-t4vask\"> </span>"), _i = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span> </span> <span class=\"step muted svelte-t4vask\"> </span> <!> <!> <span class=\"faint when svelte-t4vask\"> </span></li>"), vi = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"pill\"> </span><span class=\"step muted svelte-t4vask\"> </span><span class=\"faint\">no report yet</span></li>"), yi = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"faint\">waiting for a step to start</span></li>"), bi = /* @__PURE__ */ X("<a class=\"svelte-t4vask\"> </a>"), xi = /* @__PURE__ */ X("<li class=\"svelte-t4vask\"><span class=\"pill\">comment</span><span class=\"step muted mono svelte-t4vask\"><!></span><span class=\"note svelte-t4vask\"> </span><span class=\"faint when svelte-t4vask\"> </span></li>"), Si = /* @__PURE__ */ X("<div class=\"item card svelte-t4vask\" data-testid=\"progress-card\"><button class=\"head svelte-t4vask\"><span class=\"dot accent\"></span> <span class=\"title svelte-t4vask\"> </span> <span class=\"mono faint id svelte-t4vask\"> </span> <span class=\"faint when svelte-t4vask\"> </span></button> <div class=\"who svelte-t4vask\"><!></div> <ul class=\"svelte-t4vask\"><!> <!></ul></div>"), Ci = /* @__PURE__ */ X("<div class=\"progress svelte-t4vask\" data-testid=\"progress\"></div>"), wi = /* @__PURE__ */ X("<div class=\"faint\">Loading…</div>"), Ti = /* @__PURE__ */ X("<div class=\"faint none svelte-t4vask\" data-testid=\"progress-empty\">No runs in progress.</div>"), Ei = /* @__PURE__ */ X("<header class=\"vh\"><h1>Attention</h1> <span class=\"sub\">What waits on a human, oldest first.</span> <span class=\"actions\"><!></span></header> <div class=\"pad\"><!> <!> <h2 class=\"st\">In progress</h2> <!></div>", 1);
function Di(e, t) {
	Be(t, !0);
	let n = /* @__PURE__ */ M(null), r = /* @__PURE__ */ M(""), i = /* @__PURE__ */ M(P([])), a = /* @__PURE__ */ M(P({})), o = /* @__PURE__ */ M(!1), s = /* @__PURE__ */ M(""), c = /* @__PURE__ */ M(P({
		state: "off",
		error: ""
	})), l = P({}), u = P({}), d = 0;
	async function f() {
		let e = Object.keys(l);
		try {
			let i = await t.api.get("api/attention"), a = new Set(i.map((e) => e.kind + ":" + e.id)), o = !1;
			for (let t of e) {
				let e = l[t];
				e && (a.has(t) ? ++e.loads >= 2 ? (u[t] = e.label, delete l[t], Hr(t)) : o = !0 : (delete l[t], Hr(t)));
			}
			N(n, i, !0), o && !d && (d = window.setTimeout(() => {
				d = 0, f();
			}, 1500)), N(r, "");
		} catch (e) {
			N(r, String(e.message || e), !0);
		}
		try {
			let e = await t.api.get("api/work?state=running&limit=6&detail=1"), n = e.detail ?? {};
			N(a, e.sessions ?? {}, !0), N(i, (e.runs ?? []).slice(0, 6).map((e) => ({
				run: e,
				steps: n[e.id]?.steps ?? [],
				comment: n[e.id]?.comment ?? null
			})), !0);
		} catch {}
		N(o, !0);
	}
	function p(e, t) {
		delete u[e], l[e] = {
			label: t,
			loads: 0
		};
	}
	rn(() => Vn(() => {
		for (let [e, t] of Br()) p(e, t);
		let e = Ur((e, t) => {
			p(e, t), f();
		});
		f(), t.api.push.status().then((e) => N(c, e, !0));
		let n = 0, r = t.api.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, f();
			}, 1e3);
		}), i = t.api.params();
		if (i.item && !i.session) {
			let [e, ...t] = i.item.split(":");
			m({
				kind: e,
				id: t.join(":")
			});
		}
		return () => {
			r(), e(), clearTimeout(n), clearTimeout(d), d = 0;
		};
	}));
	async function m(e) {
		let n = e.kind + ":" + e.id;
		N(s, n);
		try {
			let r = await t.api.post("api/attention/chat", {
				kind: e.kind,
				id: e.id,
				title: e.title || "",
				why: e.why || ""
			});
			t.api.openSession(r.session, { item: n });
		} catch (r) {
			t.api.toast("Could not open a chat for " + n + ": " + String(r.message || r)), [
				"run",
				"ticket",
				"question",
				"decision",
				"stale",
				"wait"
			].includes(e.kind) && t.api.openTicket(e.id);
		} finally {
			N(s, "");
		}
	}
	let h = /* @__PURE__ */ M("");
	async function g(e, n) {
		let r = e.kind + ":" + e.id;
		N(h, r);
		try {
			await Lr(t.api, n) && Vr(r, n.label);
		} finally {
			N(h, "");
		}
	}
	let _ = {
		unsupported: "This browser has no Web Push (on iOS, add Desk to the home screen first).",
		unavailable: "Desk runs without --push.",
		denied: "Notifications are blocked in the browser settings.",
		off: "Get a notification for each new attention item.",
		on: "Notifications are on for this browser."
	};
	async function v() {
		N(c, Y(c).state === "on" ? await t.api.push.unsubscribe() : await t.api.push.subscribe(), !0);
	}
	async function y() {
		N(c, await t.api.push.test(), !0), Y(c).error || t.api.toast("Test notification sent");
	}
	function b(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	let x = (e, t) => typeof e?.[t] == "string" ? JSON.parse(e[t] || "{}") : e?.[t] || {};
	function S(e) {
		return e.map((e) => ({
			s: e,
			r: x(e, "meta").report
		})).filter((e) => e.r && (e.r.note || e.r.state)).sort((e, t) => String(t.r.at).localeCompare(String(e.r.at)));
	}
	let ee = (e) => e === "failed" || e === "blocked" ? "bad" : e === "waiting" || e === "dropped" ? "warn" : e === "ready" ? "ok" : "accent", C = (e) => !!e && /^[a-z][a-z0-9]*-[0-9a-z]{6,}$/i.test(e), te = (e) => e?.status === "in_progress" && e.holder ? Y(a)[e.holder] ?? { id: e.holder } : null, ne = (e) => t.api.openSession(e, { readonly: !0 });
	var re = Ei(), ie = Gt(re), ae = R(I(ie), 4), oe = I(ae), se = (e) => {
		var t = $r(), n = Gt(t), r = (e) => {
			var t = Qr();
			Gn("click", t, y), Z(e, t);
		};
		$(n, (e) => {
			Y(c).state === "on" && e(r);
		});
		var i = R(n, 2);
		let a;
		var o = L(i, !0);
		z(() => {
			kr(i, "title", _[Y(c).state]), a = Tr(i, 1, "", null, a, { on: Y(c).state === "on" }), Q(o, Y(c).state === "on" ? "Notifications on" : "Notify me");
		}), Gn("click", i, v), Z(e, t);
	}, ce = (e) => {
		var t = ei(), n = L(t, !0);
		z(() => {
			kr(t, "title", _[Y(c).state]), Q(n, Y(c).state === "unavailable" ? "" : "no notifications");
		}), Z(e, t);
	};
	$(oe, (e) => {
		Y(c).state === "on" || Y(c).state === "off" ? e(se) : e(ce, -1);
	}), D(ae), D(ie);
	var le = R(ie, 2), ue = I(le), de = (e) => {
		var t = ti(), n = L(t, !0);
		z(() => Q(n, Y(c).error)), Z(e, t);
	};
	$(ue, (e) => {
		Y(c).error && e(de);
	});
	var fe = R(ue, 2), pe = (e) => {
		var t = ti(), n = L(t, !0);
		z(() => Q(n, Y(r))), Z(e, t);
	}, me = (e) => {
		Z(e, ni());
	}, he = (e) => {
		Z(e, ri());
	}, ge = (e) => {
		var r = fi(), i = I(r), a = R(I(i));
		_r(a, 21, () => Y(n), (e) => e.kind + e.id, (e, n) => {
			let r = /* @__PURE__ */ A(() => Y(n).kind + ":" + Y(n).id), i = /* @__PURE__ */ A(() => l[Y(r)]);
			var a = di();
			let o;
			var c = I(a), d = L(I(c), !0);
			D(c);
			var f = R(c), p = I(f, !0), _ = R(p), v = I(_, !0), y = R(v), x = (e) => {
				var t = ai(), r = R(Gt(t)), i = (e) => {
					var t = ii(), r = L(t, !0);
					z(() => {
						kr(t, "href", Y(n).url), Q(r, Y(n).ref);
					}), Gn("click", t, (e) => e.stopPropagation()), Z(e, t);
				}, a = /* @__PURE__ */ A(() => Y(n).url && Wr(Y(n).url)), o = (e) => {
					var t = er();
					z(() => Q(t, Y(n).ref)), Z(e, t);
				};
				$(r, (e) => {
					Y(a) ? e(i) : e(o, -1);
				}), Z(e, t);
			};
			$(y, (e) => {
				Y(n).ref && e(x);
			}), D(_);
			var S = R(_, 2), ee = (e) => {
				var t = oi(), r = L(t, !0);
				z(() => Q(r, Y(n).summary)), Z(e, t);
			};
			$(S, (e) => {
				Y(n).summary && e(ee);
			});
			var C = R(S, 2), te = I(C), ne = (e) => {
				var t = si(), n = L(t);
				z(() => Q(n, `${Y(i).label ?? ""}…`)), Z(e, t);
			}, re = (e) => {
				var i = ui(), a = Gt(i), o = (e) => {
					var t = ci(), n = L(t);
					z(() => Q(n, `still listed after ${u[Y(r)] ?? ""}`)), Z(e, t);
				};
				$(a, (e) => {
					u[Y(r)] && e(o);
				});
				var c = R(a, 2);
				_r(c, 17, () => Ir(t.api, Y(n).kind, Y(n).id), (e) => e.label, (e, t) => {
					var i = li(), a = L(i, !0);
					z(() => {
						i.disabled = Y(h) === Y(r), Q(a, Y(t).label);
					}), Gn("click", i, () => g(Y(n), Y(t))), Z(e, i);
				});
				var l = R(c, 2);
				z(() => l.disabled = Y(s) === Y(r)), Gn("click", l, () => m(Y(n))), Z(e, i);
			};
			$(te, (e) => {
				Y(i) ? e(ne) : e(re, -1);
			}), D(C), D(f);
			var ie = R(f), ae = L(ie, !0), oe = L(R(ie), !0);
			D(a), z((e) => {
				o = Tr(a, 1, "click svelte-t4vask", null, o, {
					busy: Y(s) === Y(r) || !!Y(i),
					pending: !!Y(i)
				}), kr(a, "data-item", Y(n).kind + ":" + Y(n).id), Q(d, Y(n).kind), Q(p, Y(n).title || Y(n).id), Q(v, Y(n).id), Q(ae, Y(n).why), Q(oe, e);
			}, [() => b(Y(n).since)]), Gn("click", a, () => m(Y(n))), Gn("click", C, (e) => e.stopPropagation()), Z(e, a);
		}), D(a), D(i), D(r), Z(e, r);
	};
	$(fe, (e) => {
		Y(r) ? e(pe) : Y(n) ? Y(n).length === 0 ? e(he, 2) : e(ge, -1) : e(me, 1);
	});
	var _e = R(fe, 4), ve = (e) => {
		var n = Ci();
		_r(n, 21, () => Y(i), ({ run: e, steps: t, comment: n }) => e.id, (e, n) => {
			let r = () => Y(n).run, i = () => Y(n).steps, a = () => Y(n).comment, o = /* @__PURE__ */ A(() => S(i())), s = /* @__PURE__ */ A(() => i().find((e) => e.status === "in_progress")), c = /* @__PURE__ */ A(() => te(Y(s)));
			var l = Si(), u = I(l), d = R(I(u), 2), f = L(d, !0), p = R(d, 2), m = L(p, !0), h = L(R(p, 2), !0);
			D(u);
			var g = R(u, 2), _ = I(g), v = (e) => {
				var t = pi(), n = Gt(t), r = L(n, !0);
				Zr(R(n, 2), {
					get h() {
						return Y(c);
					},
					onopen: ne
				}), z((e) => Q(r, e), [() => x(Y(s), "fields").step || Y(s).title]), Z(e, t);
			}, y = (e) => {
				Z(e, mi());
			};
			$(_, (e) => {
				Y(c) ? e(v) : e(y, -1);
			}), D(g);
			var re = R(g, 2), ie = I(re);
			_r(ie, 17, () => Y(o).slice(0, 3), ({ s: e, r: t }) => e.id, (e, t) => {
				let n = () => Y(t).s, r = () => Y(t).r, i = /* @__PURE__ */ A(() => te(n()));
				var a = _i(), o = I(a), s = L(o, !0), c = R(o, 2), l = L(c, !0), u = R(c, 2), d = (e) => {
					Zr(e, {
						get h() {
							return Y(i);
						},
						onopen: ne
					});
				}, f = (e) => {
					var t = hi(), n = L(t, !0);
					z(() => {
						kr(t, "href", "?session=" + r().by), Q(n, r().by);
					}), Gn("click", t, (e) => {
						e.preventDefault(), ne(r().by);
					}), Z(e, t);
				}, p = /* @__PURE__ */ A(() => C(r().by));
				$(u, (e) => {
					Y(i) ? e(d) : Y(p) && e(f, 1);
				});
				var m = R(u, 2), h = (e) => {
					var t = gi(), n = L(t, !0);
					z(() => Q(n, r().note)), Z(e, t);
				};
				$(m, (e) => {
					r().note && e(h);
				});
				var g = L(R(m, 2), !0);
				D(a), z((e, t, n) => {
					Tr(o, 1, `pill ${e ?? ""}`, "svelte-t4vask"), Q(s, r().state), Q(l, t), Q(g, n);
				}, [
					() => ee(r().state),
					() => x(n(), "fields").step || n().title,
					() => b(r().at)
				]), Z(e, a);
			}, (e) => {
				var t = tr(), n = Gt(t), r = (e) => {
					var t = tr(), n = Gt(t), r = (e) => {
						var t = vi(), n = I(t), r = L(n, !0), i = L(R(n), !0);
						Se(), D(t), z((e) => {
							Q(r, Y(s).status), Q(i, e);
						}, [() => x(Y(s), "fields").step || Y(s).title]), Z(e, t);
					};
					$(n, (e) => {
						Y(c) || e(r);
					}), Z(e, t);
				}, i = (e) => {
					Z(e, yi());
				};
				$(n, (e) => {
					Y(s) ? e(r) : e(i, -1);
				}), Z(e, t);
			});
			var ae = R(ie, 2), oe = (e) => {
				var n = xi(), r = R(I(n)), i = I(r), o = (e) => {
					var n = bi(), r = L(n, !0);
					z(() => {
						kr(n, "href", "?session=" + a().author), Q(r, a().author);
					}), Gn("click", n, (e) => {
						e.preventDefault(), t.api.openSession(a().author);
					}), Z(e, n);
				}, s = /* @__PURE__ */ A(() => C(a().author)), c = (e) => {
					var t = er();
					z(() => Q(t, a().author)), Z(e, t);
				};
				$(i, (e) => {
					Y(s) ? e(o) : e(c, -1);
				}), D(r);
				var l = R(r), u = L(l, !0), d = L(R(l), !0);
				D(n), z((e) => {
					Q(u, a().body), Q(d, e);
				}, [() => b(a().created_at)]), Z(e, n);
			};
			$(ae, (e) => {
				a() && e(oe);
			}), D(re), D(l), z((e) => {
				kr(l, "data-run", r().id), Q(f, r().title), Q(m, r().id), Q(h, e);
			}, [() => b(r().updated_at)]), Gn("click", u, () => t.api.openTicket(r().id)), Z(e, l);
		}), D(n), Z(e, n);
	}, ye = (e) => {
		Z(e, wi());
	}, w = (e) => {
		Z(e, Ti());
	};
	$(_e, (e) => {
		Y(i).length ? e(ve) : Y(o) ? e(w, -1) : e(ye, 1);
	}), D(le), Z(e, re), Ve();
}
Kn(["click"]);
// plugins/attention/index.svelte.js
var Oi = new URL("./style.css", import.meta.url).href, ki = class extends HTMLElement {
	#e = /* @__PURE__ */ M(P({
		api: null,
		context: {},
		config: {}
	}));
	get props() {
		return Y(this.#e);
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
		if (!document.querySelector(`link[href="${Oi}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = Oi, document.head.append(e);
		}
		this.style.display = "contents", this.app = cr(Di, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && fr(this.app), this.app = null;
	}
};
customElements.get("desk-attention-view") || customElements.define("desk-attention-view", ki);
