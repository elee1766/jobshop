// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = Array.isArray, i = Array.prototype.indexOf, a = Array.prototype.includes, o = Array.from, s = Object.defineProperty, c = Object.getOwnPropertyDescriptor, l = Object.getOwnPropertyDescriptors, u = Object.prototype, d = Array.prototype, f = Object.getPrototypeOf, p = Object.isExtensible, m = () => {};
function ee(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function te() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var h = 1024, g = 2048, _ = 4096, ne = 8192, re = 16384, ie = 32768, ae = 1 << 25, oe = 65536, se = 1 << 19, ce = 1 << 20, le = 1 << 21, ue = 1 << 22, de = 1 << 23, fe = Symbol("$state"), pe = Symbol("component"), me = Symbol(""), he = Symbol("attributes"), ge = Symbol("class"), _e = Symbol("style"), ve = Symbol("text"), ye = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), be = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function xe() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function Se(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function Ce() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var v = !1;
function we(e) {
	v = e;
}
var y;
function b(t) {
	if (t === null) throw Se(), e;
	return y = t;
}
function Te() {
	return b(/* @__PURE__ */ M(y));
}
function Ee(t) {
	if (v) {
		if (/* @__PURE__ */ M(y) !== null) throw Se(), e;
		y = t;
	}
}
function De(e = 1) {
	if (v) {
		for (var t = e, n = y; t--;) n = /* @__PURE__ */ M(n);
		y = n;
	}
}
function Oe(e = !0) {
	for (var t = 0, n = y;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ M(n);
		e && n.remove(), n = i;
	}
}
function ke(t) {
	if (!t || t.nodeType !== 8) throw Se(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Ae(e) {
	return e === this.v;
}
function je(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Me(e) {
	return !je(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function Ne() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Pe(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Fe() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Ie(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Le() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Re() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function ze() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Be() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Ve() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var x = null;
function He(e) {
	x = e;
}
function Ue(e, t = !1, n) {
	x = {
		p: x,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function We(e) {
	var t = x, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) en(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, x = t.p, Ge(e);
}
function Ge(e = {}) {
	return s(e, pe, { value: !0 }), e;
}
function Ke() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var qe = [];
function Je() {
	var e = qe;
	qe = [], ee(e);
}
function S(e) {
	if (qe.length === 0 && !gt) {
		var t = qe;
		queueMicrotask(() => {
			t === qe && Je();
		});
	}
	qe.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Ye = ~(g | _ | h);
function C(e, t) {
	e.f = e.f & Ye | t;
}
function Xe(e) {
	e.f & 512 || e.deps === null ? C(e, h) : C(e, _);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Ze(e, t) {
	e.f & 2048 ? t.set(e, g) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, _), C(e, h);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Qe(e) {
	var t = B, n = U;
	H(null), W(null);
	try {
		return e();
	} finally {
		H(t), W(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function $e(e, t, n, r) {
	let i = Ke() ? rt : st;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = U, c = et(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				F(e, s);
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
		Promise.all(n.map((e) => /* @__PURE__ */ at(e))).then(u).catch((e) => F(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), tt());
	}) : f();
}
function et() {
	var e = U, t = B, n = x, r = w;
	return function(i = !0) {
		W(e), H(t), He(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function tt(e = !0) {
	W(null), H(null), He(null), e && w?.deactivate();
}
function nt() {
	var e = U, t = e.b, n = w, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function rt(e) {
	var n = 2 | g;
	return U !== null && (U.f |= se), {
		ctx: x,
		deps: null,
		effects: null,
		equals: Ae,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: U,
		ac: null
	};
}
var it = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function at(e, n, r) {
	let i = U;
	i === null && Ne();
	var a = void 0, o = At(t), s = !B, c = /* @__PURE__ */ new Set();
	return nn(() => {
		var t = U, n = te();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== ye && n.reject(e);
			}).finally(tt);
		} catch (e) {
			n.reject(e), tt();
		}
		var r = w;
		if (s) {
			if (t.f & 32768) var l = nt();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(it);
			else for (let e of c.values()) e.reject(it);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== it && (r.activate(), t ? (o.f |= de, Mt(o, t)) : (o.f & 8388608 && (o.f ^= de), Mt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), Qt(() => {
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
	return yn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function st(e) {
	let t = /* @__PURE__ */ rt(e);
	return t.equals = Me, t;
}
function ct(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) R(t[n]);
	}
}
function lt(e) {
	var n, r = U, i = e.parent;
	if (!z && i !== null && e.v !== t && i.f & 24576) return xe(), e.v;
	W(i);
	try {
		ct(e), n = En(e);
	} finally {
		W(r);
	}
	return n;
}
function ut(e) {
	var t = lt(e);
	!e.equals(t) && (e.wv = Cn(), (!w?.is_fork || e.deps === null) && (w === null ? e.v = t : (w.capture(e, t, !0), mt?.capture(e, t, !0)), e.deps === null)) ? C(e, h) : z || (T === null ? Xe(e) : (Zt() || w?.is_fork) && T.set(e, t));
}
function dt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Qe(() => {
		t.ac.abort(ye), t.ac = null;
	}), t.fn !== null && (t.teardown = m), kn(t, 0), cn(t));
}
function ft(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && An(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var pt = null, w = null, mt = null, T = null, ht = null, gt = !1, _t = !1, vt = null, yt = null, bt = 0, xt = 1, St = class e {
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
			for (var r of n.d) C(r, g), t(r);
			for (r of n.m) C(r, _), t(r);
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
					t.f ^= h;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) C(e, t), this.schedule(e);
		this.apply();
		for (var t = vt = [], n = [], r = yt = []; this.#a.length > 0;) {
			bt++ > 1e3 && (this.#b(), Ct());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Dt(e), this.#d() || this.discard(), t;
			}
		}
		if (w = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (vt = null, yt = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) Et(e, t);
			r.length > 0 && w.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), mt = this, wt(n), wt(t), mt = null, this.#i?.resolve();
			var o = w;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (D.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= h;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= h : i & 4 ? t.push(r) : wn(r) && (i & 16 && this.add_dirty_reaction(r, _), An(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), C(i, g), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), w = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Ze(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), T?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		w = this;
	}
	deactivate() {
		w = null, T = null;
	}
	flush() {
		try {
			_t = !0, w = this, this.#m();
		} finally {
			bt = 0, ht = null, vt = null, yt = null, _t = !1, w = null, T = null, D.clear();
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
		this.#u || (this.#u = !0, S(() => {
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
		return (this.#i ??= te()).promise;
	}
	static ensure() {
		if (w === null) {
			let t = w = new e();
			!_t && S(() => {
				t.started || t.flush();
			});
		}
		return w;
	}
	apply() {
		T = null;
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
		Le();
	} catch (e) {
		F(e, ht);
	}
}
var E = null;
function wt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && wn(r) && (E = /* @__PURE__ */ new Set(), An(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && dn(r), E?.size > 0)) {
				D.clear();
				for (let e of E) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) E.has(n) && (E.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || An(n);
					}
				}
				E.clear();
			}
		}
		E = null;
	}
}
function Tt(e) {
	w.schedule(e);
}
function Et(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), C(e, h);
		for (var n = e.first; n !== null;) Et(n, t), n = n.next;
	}
}
function Dt(e) {
	C(e, h);
	for (var t = e.first; t !== null;) Dt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Ot = /* @__PURE__ */ new Set(), D = /* @__PURE__ */ new Map(), kt = !1;
function At(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Ae,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function O(e, t) {
	let n = At(e, t);
	return yn(n), n;
}
function k(e, t, n = !1) {
	return B !== null && (!V || B.f & 131072) && Ke() && B.f & 4325394 && (G === null || !G.has(e)) && Be(), Mt(e, n ? It(t) : t, yt);
}
var A = null, jt = 0;
function Mt(e, t, n = null) {
	if (!e.equals(t)) {
		z ? D.set(e, t) : D.has(e) || D.set(e, e.v);
		var r = St.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && lt(t), T === null && Xe(t);
		}
		e.wv = Cn(), A = null, jt = 0, Ft(e, g, n), A = null, Ke() && U !== null && U.f & 1024 && !(U.f & 96) && (J === null ? bn([e]) : J.push(e)), !r.is_fork && Ot.size > 0 && !kt && Nt();
	}
	return t;
}
function Nt() {
	kt = !1;
	for (let e of Ot) {
		e.f & 1024 && C(e, _);
		let t;
		try {
			t = wn(e);
		} catch {
			t = !0;
		}
		t && An(e);
	}
	Ot.clear();
}
function Pt(e) {
	k(e, e.v + 1);
}
function Ft(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ke(), a = r.length;
		if (jt += a, jt > 1e5 && A === null && (A = /* @__PURE__ */ new Set()), A !== null) {
			if (A.has(e)) return;
			A.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & g) === 0;
				if (l && C(s, t), c & 131072) Ot.add(s);
				else if (c & 2) {
					var u = s;
					T?.delete(u), Ft(u, _, n);
				} else if (l) {
					var d = s;
					c & 16 && E !== null && E.add(d), n === null ? Tt(d) : n.push(d);
				}
			}
		}
	}
}
function It(e) {
	if (typeof e != "object" || !e || fe in e || pe in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ O(0), s = null, l = X, p = (e) => {
		if (X === l) return e();
		var t = B, n = X;
		H(null), Sn(l);
		var r = e();
		return H(t), Sn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ O(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Re();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ O(n.value, s);
				return i.set(t, e), e;
			}) : k(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ O(t, s));
					i.set(n, e), Pt(o);
				}
			} else k(r, t), Pt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === fe) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ O(It(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
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
			if (n === fe) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (U !== null || B !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ O(a ? It(e[n]) : t, s)), i.set(n, r)), Z(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ O(t, s)), i.set(f + "", m)) : k(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ O(void 0, s)), k(u, It(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var ee = p(() => It(r));
				k(u, ee);
			}
			var te = Reflect.getOwnPropertyDescriptor(e, n);
			if (te?.set && te.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var h = i.get("length"), g = Number(n);
					Number.isInteger(g) && g >= h.v && k(h, g + 1);
				}
				Pt(o);
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
			ze();
		}
	});
}
var Lt, Rt, zt, Bt;
function Vt() {
	if (Lt === void 0) {
		Lt = window, Rt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		zt = c(t, "firstChild").get, Bt = c(t, "nextSibling").get, p(e) && (e[ge] = void 0, e[he] = null, e[_e] = void 0, e.__e = void 0), p(n) && (n[ve] = void 0);
	}
}
function j(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Ht(e) {
	return zt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function M(e) {
	return Bt.call(e);
}
function Ut(e, t) {
	if (!v) return /* @__PURE__ */ Ht(e);
	var n = /* @__PURE__ */ Ht(y);
	if (n === null) n = y.appendChild(j());
	else if (t && n.nodeType !== 3) {
		var r = j();
		return n?.before(r), b(r), r;
	}
	return t && qt(n), b(n), n;
}
function Wt(e, t = !1) {
	if (!v) {
		var n = /* @__PURE__ */ Ht(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ M(n) : n;
	}
	if (t) {
		if (y?.nodeType !== 3) {
			var r = j();
			return y?.before(r), b(r), r;
		}
		qt(y);
	}
	return y;
}
function N(e, t = !1) {
	if (!v) return /* @__PURE__ */ Ht(e);
	var n = Ut(e, t);
	return Ee(e), n;
}
function P(e, t = 1, n = !1) {
	let r = v ? y : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ M(r);
	if (!v) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = j();
			return r === null ? i?.after(a) : r.before(a), b(a), a;
		}
		qt(r);
	}
	return b(r), r;
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
	var t = U;
	if (t === null) return B.f |= de, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	F(e, t);
}
function F(t, n) {
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
	U === null && (B === null && Ie(e), Fe()), z && Pe(e);
}
function Xt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function I(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= ne);
	var r = {
		ctx: x,
		deps: null,
		nodes: null,
		f: e | g | 512,
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
	w?.register_created_effect(r);
	var i = r;
	if (e & 4) vt === null ? St.ensure().schedule(r) : vt.push(r);
	else if (t !== null) {
		try {
			An(r);
		} catch (e) {
			throw R(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= oe));
	}
	if (i !== null && (i.parent = n, n !== null && Xt(i, n), B !== null && B.f & 2 && !(e & 64))) {
		var a = B;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Zt() {
	return B !== null && !V;
}
function Qt(e) {
	let t = I(8, null);
	return C(t, h), t.teardown = e, t;
}
function $t(e) {
	Yt("$effect");
	var t = U.f;
	if (!B && t & 32 && x !== null && !x.i) {
		var n = x;
		(n.e ??= []).push(e);
	} else return en(e);
}
function en(e) {
	return I(4 | ce, e);
}
function tn(e) {
	St.ensure();
	let t = I(64 | se, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? fn(t, () => {
			R(t), n(void 0);
		}) : (R(t), n(void 0));
	});
}
function nn(e) {
	return I(ue | se, e);
}
function rn(e, t = 0) {
	return I(8 | t, e);
}
function an(e, t = [], n = [], r = []) {
	$e(r, t, n, (t) => {
		I(8, () => {
			e(...t.map(Z));
		});
	});
}
function on(e, t = 0) {
	return I(16 | t, e);
}
function L(e) {
	return I(32 | se, e);
}
function sn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = z, r = B;
		vn(!0), H(null);
		try {
			t.call(null);
		} catch (t) {
			F(t, e.parent);
		} finally {
			vn(n), H(r);
		}
	}
}
function cn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Qe(() => {
			e.abort(ye);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : R(n, t), n = r;
	}
}
function ln(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || R(t), t = n;
	}
}
function R(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (un(e.nodes.start, e.nodes.end), n = !0), e.f |= ae, cn(e, t && !n), kn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	sn(e), e.f ^= ae, e.f |= re;
	var i = e.parent;
	i !== null && i.first !== null && dn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function un(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ M(e);
		e.remove(), e = n;
	}
}
function dn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function fn(e, t, n = !0) {
	var r = [];
	e.f |= 256, pn(e, r, !0);
	var i = () => {
		n && R(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function pn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= ne;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				pn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function mn(e) {
	e.f &= -257, hn(e, !0);
}
function hn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= ne, e.f & 1024 || (C(e, g), St.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			hn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function gn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ M(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var _n = null, z = !1;
function vn(e) {
	z = e;
}
var B = null, V = !1;
function H(e) {
	B = e;
}
var U = null;
function W(e) {
	U = e;
}
var G = null;
function yn(e) {
	B !== null && (B.f & 2097152 || B.f & 2) && (G ??= /* @__PURE__ */ new Set()).add(e);
}
var K = null, q = 0, J = null;
function bn(e) {
	J = e;
}
var xn = 1, Y = 0, X = Y;
function Sn(e) {
	X = e;
}
function Cn() {
	return ++xn;
}
function wn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (wn(a) && ut(a), a.wv > e.wv) return !0;
		}
		t & 512 && T === null && C(e, h);
	}
	return !1;
}
function Tn(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(G !== null && G.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Tn(a, t, !1) : t === a && (n ? C(a, g) : a.f & 1024 && C(a, _), Tt(a));
	}
}
function En(e) {
	var t = K, n = q, r = J, i = B, a = G, o = x, s = V, c = X, l = e.f;
	K = null, q = 0, J = null, B = l & 96 ? null : e, G = null, He(e.ctx), V = !1, X = ++Y, e.ac !== null && (Qe(() => {
		e.ac.abort(ye);
	}), e.ac = null);
	try {
		e.f |= le;
		var u = e.fn, d = u();
		e.f |= ie;
		var f = Dn(e);
		if (Ke() && J !== null && !V && f !== null && !(e.f & 6146)) for (var p = 0; p < J.length; p++) Tn(J[p], e);
		if (i !== null && i !== e) {
			if (Y++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Y;
			if (t !== null) for (let e of t) e.rv = Y;
			J !== null && (r === null ? r = J : r.push(...J));
		}
		return e.f & 8388608 && (e.f ^= de), d;
	} catch (t) {
		return Dn(e), Jt(t);
	} finally {
		e.f ^= le, K = t, q = n, J = r, B = i, G = a, He(o), V = s, X = c;
	}
}
function Dn(e) {
	var t = e.deps, n = w?.is_fork;
	if (K !== null) {
		var r;
		if (n || kn(e, q), t !== null && q > 0) for (t.length = q + K.length, r = 0; r < K.length; r++) t[q + r] = K[r];
		else e.deps = t = K;
		if (Zt() && e.f & 512) for (r = q; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && q < t.length && (kn(e, q), t.length = q);
	return t;
}
function On(e, n) {
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
		c.f & 512 && (c.f ^= 512), c.v !== t && Xe(c), c.ac !== null && Qe(() => {
			c.ac.abort(ye), c.ac = null, C(c, g);
		}), dt(c), kn(c, 0);
	}
}
function kn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) On(e, n[r]);
}
function An(e) {
	var t = e.f;
	if (!(t & 16384)) {
		C(e, h);
		var n = U;
		U = e;
		try {
			t & 16777232 ? ln(e) : cn(e), sn(e);
			var r = En(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = xn;
		} finally {
			U = n;
		}
	}
}
function Z(e) {
	var t = !!(e.f & 2);
	if (_n?.add(e), B !== null && !V && !(U !== null && U.f & 16384) && (G === null || !G.has(e))) {
		var n = B.deps;
		if (B.f & 2097152) e.rv < Y && (e.rv = Y, K === null && n !== null && n[q] === e ? q++ : K === null ? K = [e] : K.push(e));
		else {
			B.deps ??= [], a.call(B.deps, e) || B.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [B] : a.call(r, B) || r.push(B);
		}
	}
	if (z && D.has(e)) return D.get(e);
	if (t) {
		var i = e;
		if (z) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Mn(i)) && (o = lt(i)), D.set(i, o), o;
		}
		var s = !(i.f & 512) && !V && B !== null && !!(B.f & 512), c = (i.f & ie) === 0;
		wn(i) && (s && (i.f |= 512), ut(i)), s && !c && (ft(i), jn(i));
	}
	if (T?.has(e)) return T.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function jn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (ft(n), jn(n));
	}
}
function Mn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (D.has(t) || t.f & 2 && Mn(t)) return !0;
	return !1;
}
function Nn(e) {
	var t = V;
	try {
		return V = !0, e();
	} finally {
		V = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Pn = Symbol("events"), Fn = /* @__PURE__ */ new Set(), In = /* @__PURE__ */ new Set(), Ln = null, Rn = !1;
function zn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Ln = e, Rn || (Rn = !0, setTimeout(() => {
		Rn = !1, Ln = null;
	}));
	var o = 0, c = Ln === e && e[Pn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Pn] = t;
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
		var d = B, f = U;
		H(null), W(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var ee = a[Pn]?.[r];
					ee == null || a.disabled && e.target !== a || ee.call(a, e);
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
			e[Pn] = t, delete e.currentTarget, H(d), W(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Bn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Vn(e) {
	return Bn?.createHTML(e) ?? e;
}
function Hn(e) {
	var t = Kt("template");
	return t.innerHTML = Vn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Un(e, t) {
	var n = U;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function Q(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (v) return Un(y, null), y;
		i === void 0 && (i = Hn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Ht(i)));
		var t = r || Rt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Ht(t), s = t.lastChild;
			Un(o, s);
		} else Un(t, t);
		return t;
	};
}
function $(e, t) {
	if (v) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = y), Te();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var Wn = ["touchstart", "touchmove"];
function Gn(e) {
	return Wn.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Kn(e) {
	let t = 0, n = At(0), r;
	return () => {
		Zt() && (Z(n), rn(() => (t === 0 && (r = Nn(() => e(() => Pt(n)))), t += 1, () => {
			S(() => {
				--t, t === 0 && (r?.(), r = void 0, Pt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var qn = oe | se;
function Jn(e, t, n, r) {
	new Yn(e, t, n, r);
}
var Yn = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = v ? y : null;
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
	#m = Kn(() => (this.#p = At(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = on(() => {
			if (v) {
				let e = ke(this.#t);
				Te();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, qn), v && (this.#e = y);
	}
	#h() {
		try {
			this.#a = L(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		S(r), t && (this.#s = L(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? Ce() : (t = !0, n && Ve(), this.#s !== null && fn(this.#s, () => {
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
					F(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (re | ae)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = L(() => e(this.#e)), S(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = j(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return L(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							F(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(w)) : this.#u === 0 && (this.#e.before(e), this.#c = null, fn(this.#o, () => {
					this.#o = null;
				}), this.#x(w));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = L(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				gn(this.#a, e);
				let t = this.#n.pending;
				this.#o = L(() => t(this.#e));
			} else this.#x(w);
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
		var t = U, n = B, r = x;
		W(this.#i), H(this.#i), He(this.#i.ctx);
		try {
			return St.ensure(), e();
		} finally {
			W(t), H(n), He(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && fn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, S(() => {
			this.#d = !1, this.#p && Mt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Z(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		w?.is_fork ? (this.#a && w.skip_effect(this.#a), this.#o && w.skip_effect(this.#o), this.#s && w.skip_effect(this.#s), w.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (R(this.#a), null), this.#o &&= (R(this.#o), null), this.#s &&= (R(this.#s), null), v && (b(this.#t), De(), b(Oe()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return L(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return F(e, this.#i.parent), null;
				}
			}));
		};
		S(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					F(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => F(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function Xn(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[ve] ??= e.nodeValue) && (e[ve] = n, e.nodeValue = `${n}`);
}
function Zn(e, t) {
	return $n(e, t);
}
var Qn = /* @__PURE__ */ new Map();
function $n(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Vt();
	var u = void 0, d = tn(() => {
		var c = r ?? n.appendChild(j());
		Jn(c, { pending: () => {} }, (n) => {
			Ue({});
			var r = x;
			if (s && (r.c = s), a && (i.$$events = a), v && Un(n, null), u = t(n, i) || Ge(), v && (U.nodes.end = y, y === null || y.nodeType !== 8 || y.data !== "]")) throw Se(), e;
			We();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Gn(r);
					for (let e of [n, document]) {
						var a = Qn.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), Qn.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, zn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Fn)), In.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = Qn.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, zn), t.delete(e), t.size === 0 && Qn.delete(r)) : t.set(e, i);
			}
			In.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return er.set(u, d), u;
}
var er = /* @__PURE__ */ new WeakMap();
function tr(e, t) {
	let n = er.get(e);
	return n ? (er.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var nr = class {
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
			if (n) mn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (mn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (R(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						gn(r, t), t.append(j()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else R(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), fn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (R(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = w, r = Gt();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = j();
				i.append(a), this.#n.set(e, {
					effect: L(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, L(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else v && (this.anchor = y), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function rr(e, t, n = !1) {
	var r;
	v && (r = y, Te());
	var i = new nr(e), a = n ? oe : 0;
	function o(e, t) {
		if (v) {
			var n = ke(r);
			if (e !== parseInt(n.substring(1))) {
				var a = Oe();
				b(a), i.anchor = a, we(!1), i.ensure(e, t), we(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	on(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var ir = Symbol("is custom element"), ar = Symbol("is html"), or = be ? "link" : "LINK";
function sr(e, t, n, r) {
	var i = cr(e);
	if (!(v && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === or))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[me] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && ur(e).has(t) ? e[t] = n : (!v || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function cr(e) {
	return e[he] ??= {
		[ir]: e.nodeName.includes("-"),
		[ar]: e.namespaceURI === n
	};
}
var lr = /* @__PURE__ */ new Map();
function ur(e) {
	var t = e.getAttribute("is") || e.nodeName, n = lr.get(t);
	if (n) return n;
	lr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// plugins/deploys/Deploys.svelte
var dr = /* @__PURE__ */ Q("<div class=\"card err svelte-wubzx8\"> </div>"), fr = /* @__PURE__ */ Q("<a target=\"_blank\" rel=\"noopener noreferrer\" class=\"mono\"> </a>"), pr = /* @__PURE__ */ Q("<span class=\"pill warn\">controller not restarted on it yet</span>"), mr = /* @__PURE__ */ Q("<a class=\"v mono svelte-wubzx8\" target=\"_blank\" rel=\"noopener noreferrer\"> </a> <div class=\"muted\"> </div> <!>", 1), hr = /* @__PURE__ */ Q("<div class=\"muted\">no self-deploy state</div>"), gr = /* @__PURE__ */ Q("<span class=\"pill\">superseded</span>"), _r = /* @__PURE__ */ Q("<span class=\"pill bad\">rolled back</span>"), vr = /* @__PURE__ */ Q("<div class=\"muted\">none</div>"), yr = /* @__PURE__ */ Q("<div class=\"grid svelte-wubzx8\" data-testid=\"deploys\"><div class=\"card cell svelte-wubzx8\"><div class=\"faint\">Controller running</div> <div class=\"v mono svelte-wubzx8\"> </div> <!></div> <div class=\"card cell svelte-wubzx8\"><div class=\"faint\">Deployed rev</div> <!></div> <div class=\"card cell svelte-wubzx8\"><div class=\"faint\">Last failed rev (rolled back)</div> <!></div></div>"), br = /* @__PURE__ */ Q("<div class=\"desk-deploys svelte-wubzx8\"><h1 class=\"svelte-wubzx8\">Deploys</h1> <!></div>");
function xr(e, t) {
	Ue(t, !0);
	let n = /* @__PURE__ */ O(null), r = /* @__PURE__ */ O("");
	async function i() {
		try {
			k(n, await t.api.get("api/deploys"), !0), k(r, "");
		} catch (e) {
			k(r, String(e.message || e), !0);
		}
	}
	$t(() => {
		if (!t.api) return;
		i();
		let e = 0, n = t.api.subscribe(() => {
			e ||= window.setTimeout(() => {
				e = 0, i();
			}, 1e3);
		});
		return () => {
			n(), clearTimeout(e);
		};
	});
	let a = (e) => e ? "https://github.com/elee1766/jobshop/commit/" + encodeURIComponent(e) : "", o = (e) => (e || "").slice(0, 7);
	function s(e) {
		if (!e || e.startsWith("0001")) return "";
		let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3);
		return t < 60 ? t + "s ago" : t < 3600 ? Math.round(t / 60) + "m ago" : t < 86400 ? Math.round(t / 3600) + "h ago" : Math.round(t / 86400) + "d ago";
	}
	var c = br(), l = P(Ut(c), 2), u = (e) => {
		var t = dr(), n = N(t, !0);
		an(() => Xn(n, Z(r))), $(e, t);
	}, d = (e) => {
		let t = /* @__PURE__ */ ot(() => Z(n)), r = /* @__PURE__ */ ot(() => Z(t).deployed && Z(t).server.commit && !Z(t).deployed.startsWith(Z(t).server.commit));
		var i = yr(), c = Ut(i), l = P(Ut(c), 2), u = N(l, !0), d = P(l, 2), f = (e) => {
			var n = fr(), r = N(n, !0);
			an((e, t) => {
				sr(n, "href", e), Xn(r, t);
			}, [() => a(Z(t).server.commit), () => o(Z(t).server.commit)]), $(e, n);
		};
		rr(d, (e) => {
			Z(t).server.commit && e(f);
		}), Ee(c);
		var p = P(c, 2), m = P(Ut(p), 2), ee = (e) => {
			var n = mr(), i = Wt(n), c = N(i, !0), l = P(i, 2), u = N(l, !0), d = P(l, 2), f = (e) => {
				$(e, pr());
			};
			rr(d, (e) => {
				Z(r) && e(f);
			}), an((e, t, n) => {
				sr(i, "href", e), Xn(c, t), Xn(u, n);
			}, [
				() => a(Z(t).deployed),
				() => o(Z(t).deployed),
				() => s(Z(t).deployed_at)
			]), $(e, n);
		}, te = (e) => {
			$(e, hr());
		};
		rr(m, (e) => {
			Z(t).deployed ? e(ee) : e(te, -1);
		}), Ee(p);
		var h = P(p, 2), g = P(Ut(h), 2), _ = (e) => {
			var n = mr(), r = Wt(n), i = N(r, !0), c = P(r, 2), l = N(c, !0), u = P(c, 2), d = (e) => {
				$(e, gr());
			}, f = (e) => {
				$(e, _r());
			};
			rr(u, (e) => {
				Z(t).failed === Z(t).deployed ? e(d) : e(f, -1);
			}), an((e, t, n) => {
				sr(r, "href", e), Xn(i, t), Xn(l, n);
			}, [
				() => a(Z(t).failed),
				() => o(Z(t).failed),
				() => s(Z(t).failed_at)
			]), $(e, n);
		}, ne = (e) => {
			$(e, vr());
		};
		rr(g, (e) => {
			Z(t).failed ? e(_) : e(ne, -1);
		}), Ee(h), Ee(i), an(() => Xn(u, Z(t).server.version)), $(e, i);
	};
	rr(l, (e) => {
		Z(r) ? e(u) : Z(n) && e(d, 1);
	}), Ee(c), $(e, c), We();
}
// plugins/deploys/index.svelte.js
var Sr = new URL("./style.css", import.meta.url).href, Cr = class extends HTMLElement {
	#e = /* @__PURE__ */ O(It({
		api: null,
		context: {},
		config: {}
	}));
	get props() {
		return Z(this.#e);
	}
	set props(e) {
		k(this.#e, e, !0);
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
		if (!document.querySelector(`link[href="${Sr}"]`)) {
			let e = document.createElement("link");
			e.rel = "stylesheet", e.href = Sr, document.head.append(e);
		}
		this.app = Zn(xr, {
			target: this,
			props: this.props
		});
	}
	disconnectedCallback() {
		this.app && tr(this.app), this.app = null;
	}
};
customElements.get("desk-deploys-view") || customElements.define("desk-deploys-view", Cr);
