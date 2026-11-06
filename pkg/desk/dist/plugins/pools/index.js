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
var g = 1024, _ = 2048, v = 4096, y = 8192, b = 16384, x = 32768, S = 1 << 25, ee = 65536, C = 1 << 19, te = 1 << 20, ne = 1 << 25, re = 1 << 21, ie = 1 << 22, ae = 1 << 23, oe = Symbol("$state"), se = Symbol("component"), ce = Symbol("attributes"), le = Symbol("class"), ue = Symbol("style"), de = Symbol("text"), fe = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}();
globalThis.document?.contentType;
function pe() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function me(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function he() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var w = !1;
function ge(e) {
	w = e;
}
var T;
function E(t) {
	if (t === null) throw me(), e;
	return T = t;
}
function _e() {
	return E(/* @__PURE__ */ Vt(T));
}
function D(t) {
	if (w) {
		if (/* @__PURE__ */ Vt(T) !== null) throw me(), e;
		T = t;
	}
}
function ve(e = 1) {
	if (w) {
		for (var t = e, n = T; t--;) n = /* @__PURE__ */ Vt(n);
		T = n;
	}
}
function ye(e = !0) {
	for (var t = 0, n = T;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Vt(n);
		e && n.remove(), n = i;
	}
}
function be(t) {
	if (!t || t.nodeType !== 8) throw me(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function xe(e) {
	return e === this.v;
}
function Se(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Ce(e) {
	return !Se(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function we() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Te(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function Ee(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function De() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Oe(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function ke() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Ae() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function je() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Me() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Ne() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var O = null;
function Pe(e) {
	O = e;
}
function Fe(e, t = !1, n) {
	O = {
		p: O,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: W,
		l: null
	};
}
function Ie(e) {
	var t = O, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) en(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, O = t.p, Le(e);
}
function Le(e = {}) {
	return o(e, se, { value: !0 }), e;
}
function Re() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var ze = [];
function Be() {
	var e = ze;
	ze = [], p(e);
}
function Ve(e) {
	if (ze.length === 0 && !lt) {
		var t = ze;
		queueMicrotask(() => {
			t === ze && Be();
		});
	}
	ze.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var He = ~(_ | v | g);
function k(e, t) {
	e.f = e.f & He | t;
}
function Ue(e) {
	e.f & 512 || e.deps === null ? k(e, g) : k(e, v);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function We(e, t) {
	e.f & 2048 ? t.set(e, _) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, v), k(e, g);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Ge(e) {
	var t = V, n = W;
	U(null), G(null);
	try {
		return e();
	} finally {
		U(t), G(n);
	}
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function Ke(e, t, n, r) {
	let i = Re() ? Xe : et;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = W, c = qe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				R(e, s);
			}
			Je();
		}
	}
	var d = Ye();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ Qe(e))).then(u).catch((e) => R(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Je());
	}) : f();
}
function qe() {
	var e = W, t = V, n = O, r = A;
	return function(i = !0) {
		G(e), U(t), Pe(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Je(e = !0) {
	G(null), U(null), Pe(null), e && A?.deactivate();
}
function Ye() {
	var e = W, t = e.b, n = A, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Xe(e) {
	var n = 2 | _;
	return W !== null && (W.f |= C), {
		ctx: O,
		deps: null,
		effects: null,
		equals: xe,
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
var Ze = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function Qe(e, n, r) {
	let i = W;
	i === null && we();
	var a = void 0, o = Tt(t), s = !V, c = /* @__PURE__ */ new Set();
	return nn(() => {
		var t = W, n = m();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== fe && n.reject(e);
			}).finally(Je);
		} catch (e) {
			n.reject(e), Je();
		}
		var r = A;
		if (s) {
			if (t.f & 32768) var l = Ye();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(Ze);
			else for (let e of c.values()) e.reject(Ze);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== Ze && (r.activate(), t ? (o.f |= ae, kt(o, t)) : (o.f & 8388608 && (o.f ^= ae), kt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), Qt(() => {
		for (let e of c) e.reject(Ze);
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
function $e(e) {
	let t = /* @__PURE__ */ Xe(e);
	return bn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function et(e) {
	let t = /* @__PURE__ */ Xe(e);
	return t.equals = Ce, t;
}
function tt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) B(t[n]);
	}
}
function nt(e) {
	var n, r = W, i = e.parent;
	if (!vn && i !== null && e.v !== t && i.f & 24576) return pe(), e.v;
	G(i);
	try {
		tt(e), n = kn(e);
	} finally {
		G(r);
	}
	return n;
}
function rt(e) {
	var t = nt(e);
	!e.equals(t) && (e.wv = En(), (!A?.is_fork || e.deps === null) && (A === null ? e.v = t : (A.capture(e, t, !0), st?.capture(e, t, !0)), e.deps === null)) ? k(e, g) : vn || (j === null ? Ue(e) : (Zt() || A?.is_fork) && j.set(e, t));
}
function it(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Ge(() => {
		t.ac.abort(fe), t.ac = null;
	}), t.fn !== null && (t.teardown = f), Mn(t, 0), cn(t));
}
function at(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Nn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ot = null, A = null, st = null, j = null, ct = null, lt = !1, ut = !1, dt = null, ft = null, pt = 0, mt = 1, ht = class e {
	id = mt++;
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
		ot === null ? ot = this : (ot.next = this, this.prev = ot), ot = this;
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
			for (var r of n.d) k(r, _), t(r);
			for (r of n.m) k(r, v), t(r);
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
		for (let [e, t] of this.#s) k(e, t), this.schedule(e);
		this.apply();
		for (var t = dt = [], n = [], r = ft = []; this.#a.length > 0;) {
			pt++ > 1e3 && (this.#b(), gt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw xt(e), this.#d() || this.discard(), t;
			}
		}
		if (A = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (dt = null, ft = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) bt(e, t);
			r.length > 0 && A.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), st = this, vt(n), vt(t), st = null, this.#i?.resolve();
			var o = A;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (Ct.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= g;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= g : i & 4 ? t.push(r) : Dn(r) && (i & 16 && this.add_dirty_reaction(r, v), Nn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), k(i, _), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), A = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) We(e[t], this.#s);
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
			ut = !0, A = this, this.#m();
		} finally {
			pt = 0, ct = null, dt = null, ft = null, ut = !1, A = null, j = null, Ct.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(Ze);
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
		this.#u || (this.#u = !0, Ve(() => {
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
		if (A === null) {
			let t = A = new e();
			!ut && Ve(() => {
				t.started || t.flush();
			});
		}
		return A;
	}
	apply() {
		j = null;
	}
	schedule(e) {
		ct = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? ot = e : t.prev = e, this.linked = !1;
		}
	}
};
function gt() {
	try {
		ke();
	} catch (e) {
		R(e, ct);
	}
}
var _t = null;
function vt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && Dn(r) && (_t = /* @__PURE__ */ new Set(), Nn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && dn(r), _t?.size > 0)) {
				Ct.clear();
				for (let e of _t) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) _t.has(n) && (_t.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Nn(n);
					}
				}
				_t.clear();
			}
		}
		_t = null;
	}
}
function yt(e) {
	A.schedule(e);
}
function bt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), k(e, g);
		for (var n = e.first; n !== null;) bt(n, t), n = n.next;
	}
}
function xt(e) {
	k(e, g);
	for (var t = e.first; t !== null;) xt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var St = /* @__PURE__ */ new Set(), Ct = /* @__PURE__ */ new Map(), wt = !1;
function Tt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: xe,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function M(e, t) {
	let n = Tt(e, t);
	return bn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Et(e, t = !1, n = !0) {
	let r = Tt(e);
	return t || (r.equals = Ce), r;
}
function N(e, t, n = !1) {
	return V !== null && (!H || V.f & 131072) && Re() && V.f & 4325394 && (K === null || !K.has(e)) && Me(), kt(e, n ? Pt(t) : t, ft);
}
var Dt = null, Ot = 0;
function kt(e, t, n = null) {
	if (!e.equals(t)) {
		vn ? Ct.set(e, t) : Ct.has(e) || Ct.set(e, e.v);
		var r = ht.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && nt(t), j === null && Ue(t);
		}
		e.wv = En(), Dt = null, Ot = 0, Nt(e, _, n), Dt = null, Re() && W !== null && W.f & 1024 && !(W.f & 96) && (Y === null ? xn([e]) : Y.push(e)), !r.is_fork && St.size > 0 && !wt && At();
	}
	return t;
}
function At() {
	wt = !1;
	for (let e of St) {
		e.f & 1024 && k(e, v);
		let t;
		try {
			t = Dn(e);
		} catch {
			t = !0;
		}
		t && Nn(e);
	}
	St.clear();
}
function jt(e, t = 1) {
	var n = X(e), r = t === 1 ? n++ : n--;
	return N(e, n), r;
}
function Mt(e) {
	N(e, e.v + 1);
}
function Nt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Re(), a = r.length;
		if (Ot += a, Ot > 1e5 && Dt === null && (Dt = /* @__PURE__ */ new Set()), Dt !== null) {
			if (Dt.has(e)) return;
			Dt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== W) {
				var l = (c & _) === 0;
				if (l && k(s, t), c & 131072) St.add(s);
				else if (c & 2) {
					var u = s;
					j?.delete(u), Nt(u, v, n);
				} else if (l) {
					var d = s;
					c & 16 && _t !== null && _t.add(d), n === null ? yt(d) : n.push(d);
				}
			}
		}
	}
}
function Pt(e) {
	if (typeof e != "object" || !e || oe in e || se in e) return e;
	let r = u(e);
	if (r !== c && r !== l) return e;
	var i = /* @__PURE__ */ new Map(), a = n(e), o = /* @__PURE__ */ M(0), d = null, f = wn, p = (e) => {
		if (wn === f) return e();
		var t = V, n = wn;
		U(null), Tn(f);
		var r = e();
		return U(t), Tn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ M(e.length, d)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ae();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ M(n.value, d);
				return i.set(t, e), e;
			}) : N(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ M(t, d));
					i.set(n, e), Mt(o);
				}
			} else N(r, t), Mt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === oe) return e;
			var o = i.get(r), c = r in n;
			if (o === void 0 && (!c || s(n, r)?.writable) && (o = p(() => /* @__PURE__ */ M(Pt(c ? n[r] : t), d)), i.set(r, o)), o !== void 0) {
				var l = X(o);
				return l === t ? void 0 : l;
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
			if (n === oe) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (W !== null || V !== null) && (!a || s(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ M(a ? Pt(e[n]) : t, d)), i.set(n, r)), X(r) === t) ? !1 : a;
		},
		set(e, n, r, c) {
			var l = i.get(n), u = n in e;
			if (a && n === "length") for (var f = r; f < l.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ M(t, d)), i.set(f + "", m)) : N(m, t);
			}
			if (l === void 0) (!u || s(e, n)?.writable) && (l = p(() => /* @__PURE__ */ M(void 0, d)), N(l, Pt(r)), i.set(n, l));
			else {
				u = l.v !== t;
				var h = p(() => Pt(r));
				N(l, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(c, r), !u) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && N(_, v + 1);
				}
				Mt(o);
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
			je();
		}
	});
}
var Ft, It, Lt, Rt;
function zt() {
	if (Ft === void 0) {
		Ft = window, It = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Lt = s(t, "firstChild").get, Rt = s(t, "nextSibling").get, d(e) && (e[le] = void 0, e[ce] = null, e[ue] = void 0, e.__e = void 0), d(n) && (n[de] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Bt(e) {
	return Lt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Vt(e) {
	return Rt.call(e);
}
function F(e, t) {
	if (!w) return /* @__PURE__ */ Bt(e);
	var n = /* @__PURE__ */ Bt(T);
	if (n === null) n = T.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), E(r), r;
	}
	return t && Kt(n), E(n), n;
}
function Ht(e, t = !1) {
	if (!w) {
		var n = /* @__PURE__ */ Bt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Vt(n) : n;
	}
	if (t) {
		if (T?.nodeType !== 3) {
			var r = P();
			return T?.before(r), E(r), r;
		}
		Kt(T);
	}
	return T;
}
function I(e, t = !1) {
	if (!w) return /* @__PURE__ */ Bt(e);
	var n = F(e, t);
	return D(e), n;
}
function L(e, t = 1, n = !1) {
	let r = w ? T : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Vt(r);
	if (!w) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), E(a), a;
		}
		Kt(r);
	}
	return E(r), r;
}
function Ut(e) {
	e.textContent = "";
}
function Wt() {
	return !1;
}
function Gt(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function Kt(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function qt(e) {
	var t = W;
	if (t === null) return V.f |= ae, e;
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
function Jt(e) {
	W === null && (V === null && Oe(e), De()), vn && Ee(e);
}
function Yt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Xt(e, t) {
	var n = W;
	n !== null && n.f & 8192 && (e |= y);
	var r = {
		ctx: O,
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
	A?.register_created_effect(r);
	var i = r;
	if (e & 4) dt === null ? ht.ensure().schedule(r) : dt.push(r);
	else if (t !== null) {
		try {
			Nn(r);
		} catch (e) {
			throw B(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= ee));
	}
	if (i !== null && (i.parent = n, n !== null && Yt(i, n), V !== null && V.f & 2 && !(e & 64))) {
		var a = V;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Zt() {
	return V !== null && !H;
}
function Qt(e) {
	let t = Xt(8, null);
	return k(t, g), t.teardown = e, t;
}
function $t(e) {
	Jt("$effect");
	var t = W.f;
	if (!V && t & 32 && O !== null && !O.i) {
		var n = O;
		(n.e ??= []).push(e);
	} else return en(e);
}
function en(e) {
	return Xt(4 | te, e);
}
function tn(e) {
	ht.ensure();
	let t = Xt(64 | C, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? fn(t, () => {
			B(t), n(void 0);
		}) : (B(t), n(void 0));
	});
}
function nn(e) {
	return Xt(ie | C, e);
}
function rn(e, t = 0) {
	return Xt(8 | t, e);
}
function an(e, t = [], n = [], r = []) {
	Ke(r, t, n, (t) => {
		Xt(8, () => {
			e(...t.map(X));
		});
	});
}
function on(e, t = 0) {
	return Xt(16 | t, e);
}
function z(e) {
	return Xt(32 | C, e);
}
function sn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = vn, r = V;
		yn(!0), U(null);
		try {
			t.call(null);
		} catch (t) {
			R(t, e.parent);
		} finally {
			yn(n), U(r);
		}
	}
}
function cn(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Ge(() => {
			e.abort(fe);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : B(n, t), n = r;
	}
}
function ln(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || B(t), t = n;
	}
}
function B(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (un(e.nodes.start, e.nodes.end), n = !0), e.f |= S, cn(e, t && !n), Mn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	sn(e), e.f ^= S, e.f |= b;
	var i = e.parent;
	i !== null && i.first !== null && dn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function un(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Vt(e);
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
		n && B(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function pn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= y;
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
		e.f ^= y, e.f & 1024 || (k(e, _), ht.ensure().schedule(e));
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
		var i = n === r ? null : /* @__PURE__ */ Vt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var _n = null, vn = !1;
function yn(e) {
	vn = e;
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
function bn(e) {
	V !== null && (V.f & 2097152 || V.f & 2) && (K ??= /* @__PURE__ */ new Set()).add(e);
}
var q = null, J = 0, Y = null;
function xn(e) {
	Y = e;
}
var Sn = 1, Cn = 0, wn = Cn;
function Tn(e) {
	wn = e;
}
function En() {
	return ++Sn;
}
function Dn(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (Dn(a) && rt(a), a.wv > e.wv) return !0;
		}
		t & 512 && j === null && k(e, g);
	}
	return !1;
}
function On(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(K !== null && K.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? On(a, t, !1) : t === a && (n ? k(a, _) : a.f & 1024 && k(a, v), yt(a));
	}
}
function kn(e) {
	var t = q, n = J, r = Y, i = V, a = K, o = O, s = H, c = wn, l = e.f;
	q = null, J = 0, Y = null, V = l & 96 ? null : e, K = null, Pe(e.ctx), H = !1, wn = ++Cn, e.ac !== null && (Ge(() => {
		e.ac.abort(fe);
	}), e.ac = null);
	try {
		e.f |= re;
		var u = e.fn, d = u();
		e.f |= x;
		var f = An(e);
		if (Re() && Y !== null && !H && f !== null && !(e.f & 6146)) for (var p = 0; p < Y.length; p++) On(Y[p], e);
		if (i !== null && i !== e) {
			if (Cn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Cn;
			if (t !== null) for (let e of t) e.rv = Cn;
			Y !== null && (r === null ? r = Y : r.push(...Y));
		}
		return e.f & 8388608 && (e.f ^= ae), d;
	} catch (t) {
		return An(e), qt(t);
	} finally {
		e.f ^= re, q = t, J = n, Y = r, V = i, K = a, Pe(o), H = s, wn = c;
	}
}
function An(e) {
	var t = e.deps, n = A?.is_fork;
	if (q !== null) {
		var r;
		if (n || Mn(e, J), t !== null && J > 0) for (t.length = J + q.length, r = 0; r < q.length; r++) t[J + r] = q[r];
		else e.deps = t = q;
		if (Zt() && e.f & 512) for (r = J; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && J < t.length && (Mn(e, J), t.length = J);
	return t;
}
function jn(e, n) {
	let a = n.reactions;
	if (a !== null) {
		var o = r.call(a, e);
		if (o !== -1) {
			var s = a.length - 1;
			s === 0 ? a = n.reactions = null : (a[o] = a[s], a.pop());
		}
	}
	if (a === null && n.f & 2 && (q === null || !i.call(q, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ue(c), c.ac !== null && Ge(() => {
			c.ac.abort(fe), c.ac = null, k(c, _);
		}), it(c), Mn(c, 0);
	}
}
function Mn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) jn(e, n[r]);
}
function Nn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		k(e, g);
		var n = W;
		W = e;
		try {
			t & 16777232 ? ln(e) : cn(e), sn(e);
			var r = kn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = Sn;
		} finally {
			W = n;
		}
	}
}
function X(e) {
	var t = !!(e.f & 2);
	if (_n?.add(e), V !== null && !H && !(W !== null && W.f & 16384) && (K === null || !K.has(e))) {
		var n = V.deps;
		if (V.f & 2097152) e.rv < Cn && (e.rv = Cn, q === null && n !== null && n[J] === e ? J++ : q === null ? q = [e] : q.push(e));
		else {
			V.deps ??= [], i.call(V.deps, e) || V.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [V] : i.call(r, V) || r.push(V);
		}
	}
	if (vn && Ct.has(e)) return Ct.get(e);
	if (t) {
		var a = e;
		if (vn) {
			var o = a.v;
			return (!(a.f & 1024) && a.reactions !== null || Fn(a)) && (o = nt(a)), Ct.set(a, o), o;
		}
		var s = !(a.f & 512) && !H && V !== null && !!(V.f & 512), c = (a.f & x) === 0;
		Dn(a) && (s && (a.f |= 512), rt(a)), s && !c && (at(a), Pn(a));
	}
	if (j?.has(e)) return j.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Pn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : i.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (at(n), Pn(n));
	}
}
function Fn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (Ct.has(t) || t.f & 2 && Fn(t)) return !0;
	return !1;
}
function In(e) {
	var t = H;
	try {
		return H = !0, e();
	} finally {
		H = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Ln = Symbol("events"), Rn = /* @__PURE__ */ new Set(), zn = /* @__PURE__ */ new Set();
function Bn(e, t, n) {
	(t[Ln] ??= {})[e] = n;
}
function Vn(e) {
	for (var t = 0; t < e.length; t++) Rn.add(e[t]);
	for (var n of zn) n(e);
}
var Hn = null, Un = !1;
function Wn(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Hn = e, Un || (Un = !0, setTimeout(() => {
		Un = !1, Hn = null;
	}));
	var s = 0, c = Hn === e && e[Ln];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Ln] = t;
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
		var d = V, f = W;
		U(null), G(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Ln]?.[r];
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
			e[Ln] = t, delete e.currentTarget, U(d), G(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var Gn = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function Kn(e) {
	return Gn?.createHTML(e) ?? e;
}
function qn(e) {
	var t = Gt("template");
	return t.innerHTML = Kn(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function Jn(e, t) {
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
		if (w) return Jn(T, null), T;
		i === void 0 && (i = qn(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Bt(i)));
		var t = r || It ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Bt(t), s = t.lastChild;
			Jn(o, s);
		} else Jn(t, t);
		return t;
	};
}
function Q(e, t) {
	if (w) {
		var n = W;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = T), _e();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var Yn = ["touchstart", "touchmove"];
function Xn(e) {
	return Yn.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function Zn(e) {
	let t = 0, n = Tt(0), r;
	return () => {
		Zt() && (X(n), rn(() => (t === 0 && (r = In(() => e(() => Mt(n)))), t += 1, () => {
			Ve(() => {
				--t, t === 0 && (r?.(), r = void 0, Mt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var Qn = ee | C;
function $n(e, t, n, r) {
	new er(e, t, n, r);
}
var er = class {
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
	#m = Zn(() => (this.#p = Tt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = W;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = W.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = on(() => {
			if (w) {
				let e = be(this.#t);
				_e();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, Qn), w && (this.#e = T);
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
		Ve(r), t && (this.#s = z(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? he() : (t = !0, n && Ne(), this.#s !== null && fn(this.#s, () => {
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
		return (this.#i.f & (b | S)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = z(() => e(this.#e)), Ve(() => {
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
				}), this.#a === null ? (this.#c = null, n && this.#x(A)) : this.#u === 0 && (this.#e.before(e), this.#c = null, fn(this.#o, () => {
					this.#o = null;
				}), this.#x(A));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = z(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				gn(this.#a, e);
				let t = this.#n.pending;
				this.#o = z(() => t(this.#e));
			} else this.#x(A);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		We(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = W, n = V, r = O;
		G(this.#i), U(this.#i), Pe(this.#i.ctx);
		try {
			return ht.ensure(), e();
		} finally {
			G(t), U(n), Pe(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && fn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ve(() => {
			this.#d = !1, this.#p && kt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), X(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		A?.is_fork ? (this.#a && A.skip_effect(this.#a), this.#o && A.skip_effect(this.#o), this.#s && A.skip_effect(this.#s), A.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (B(this.#a), null), this.#o &&= (B(this.#o), null), this.#s &&= (B(this.#s), null), w && (E(this.#t), ve(), E(ye()));
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
		Ve(() => {
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
	n !== (e[de] ??= e.nodeValue) && (e[de] = n, e.nodeValue = `${n}`);
}
function tr(e, t) {
	return rr(e, t);
}
var nr = /* @__PURE__ */ new Map();
function rr(t, { target: n, anchor: r, props: i = {}, events: o, context: s, intro: c = !0, transformError: l }) {
	zt();
	var u = void 0, d = tn(() => {
		var c = r ?? n.appendChild(P());
		$n(c, { pending: () => {} }, (n) => {
			Fe({});
			var r = O;
			if (s && (r.c = s), o && (i.$$events = o), w && Jn(n, null), u = t(n, i) || Le(), w && (W.nodes.end = T, T === null || T.nodeType !== 8 || T.data !== "]")) throw me(), e;
			Ie();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = Xn(r);
					for (let e of [n, document]) {
						var a = nr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), nr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, Wn, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(a(Rn)), zn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = nr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, Wn), t.delete(e), t.size === 0 && nr.delete(r)) : t.set(e, i);
			}
			zn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return ir.set(u, d), u;
}
var ir = /* @__PURE__ */ new WeakMap();
function ar(e, t) {
	let n = ir.get(e);
	return n ? (ir.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var or = class {
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
				r && (B(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						gn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else B(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), fn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (B(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = A, r = Wt();
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
		} else w && (this.anchor = T), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function sr(e, t, n = !1) {
	var r;
	w && (r = T, _e());
	var i = new or(e), a = n ? ee : 0;
	function o(e, t) {
		if (w) {
			var n = be(r);
			if (e !== parseInt(n.substring(1))) {
				var a = ye();
				E(a), i.anchor = a, ge(!1), i.ensure(e, t), ge(!0);
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
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function cr(e, t, n) {
	for (var r = [], i = t.length, o, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		fn(n, () => {
			if (o) {
				if (o.pending.delete(n), o.done.add(n), o.pending.size === 0) {
					var t = e.outrogroups;
					lr(e, a(o.done)), t.delete(o), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			Ut(d), d.append(u), e.items.clear();
		}
		lr(e, t, !l);
	} else o = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(o);
}
function lr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= ne, gn(a, document.createDocumentFragment())) : B(t[i], n);
	}
}
var ur;
function dr(e, t, r, i, o, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = w ? E(/* @__PURE__ */ Bt(u)) : u.appendChild(P());
	}
	w && _e();
	var d = null, f = /* @__PURE__ */ et(() => {
		var e = r();
		return n(e) ? e : e == null ? [] : a(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = X(f);
			_.fallback = d, pr(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= ne, hr(d, null, c)) : mn(d) : fn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: on(() => {
			var e = X(f), n = e.length;
			let a = !1;
			w && be(c) === "[!" != (n === 0) && (c = ye(), E(c), ge(!1), a = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = A, v = Wt(), y = 0; y < n; y += 1) {
				w && T.nodeType === 8 && T.data === "]" && (c = T, a = !0, ge(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && kt(S.v, b), S.i && kt(S.i, y), v && _.unskip_effect(S.e)) : (S = mr(l, m ? c : ur ??= P(), b, x, y, o, t, r), m || (S.e.f |= ne), l.set(x, S)), u.add(x);
			}
			if (n === 0 && s && !d && (m ? d = z(() => s(c)) : (d = z(() => s(ur ??= P())), d.f |= ne)), n > u.size && Te("", "", ""), w && n > 0 && E(ye()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			a && ge(!0), X(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, w && (c = T);
}
function fr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function pr(e, t, n, r, i) {
	var o = !!(r & 8), s = t.length, c = e.items, l = fr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (o) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (mn(_), o && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= ne, _ === l) hr(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), gr(e, d, _), gr(e, _, y), hr(_, y, n), d = _, p = [], m = [], l = fr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) hr(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					gr(e, S.prev, ee.next), gr(e, d, S), gr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), hr(_, l, n), gr(e, _.prev, _.next), gr(e, _, d === null ? e.effect.first : d.next), gr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = fr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = fr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (lr(e, a(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = fr(l.next);
		var te = C.length;
		if (te > 0) {
			var re = r & 4 && s === 0 ? n : null;
			if (o) {
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.fix();
			}
			cr(e, C, re);
		}
	}
	o && Ve(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function mr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Tt(n) : /* @__PURE__ */ Et(n, !1, !1) : null, l = o & 2 ? Tt(i) : null;
	return {
		v: c,
		i: l,
		e: z(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function hr(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Vt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function gr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/shared/attributes.js
var _r = [..." 	\n\r\f\xA0\v﻿"];
function vr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || _r.includes(r[o - 1])) && (s === r.length || _r.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function yr(e, t, n, r, i, a) {
	var o = e[le];
	if (w || o !== n || o === void 0) {
		var s = vr(n, r, a);
		w && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[le] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// kit/plugin.svelte.js
function br(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ M(Pt({
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
			this.style.display = "contents", this.app ||= tr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && ar(this.app), this.app = null;
		}
	});
}
function xr(e, t) {
	let n = Pt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			jt(r);
		}
	}), r = /* @__PURE__ */ M(0);
	$t(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, jt(r);
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
	return $t(() => {
		let n = e(), c = t();
		X(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
var Sr = (e) => String(e?.message || e);
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// kit/format.js
function Cr(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
// plugins/pools/Pools.svelte
var wr = /* @__PURE__ */ Z("<div class=\"card err\"> </div>"), Tr = /* @__PURE__ */ Z("<span class=\"faint\"> </span>"), Er = /* @__PURE__ */ Z("<span class=\"pill warn gap svelte-1td2jxo\">drained</span>"), Dr = /* @__PURE__ */ Z("<span class=\"pill accent\"> </span>"), Or = /* @__PURE__ */ Z("<span class=\"muted\">auto</span>"), kr = /* @__PURE__ */ Z("<button class=\"svelte-1td2jxo\">Auto</button>"), Ar = /* @__PURE__ */ Z("<tr><td class=\"mono c-id\"> <!></td><td data-label=\"desired\"> </td><td data-label=\"open\"> </td><td data-label=\"working\"> </td><td data-label=\"idle\"> </td><td data-label=\"cordoned\"> </td><td class=\"muted\" data-label=\"min/max\"> </td><td class=\"c-st\"><!></td><td class=\"c-act acts svelte-1td2jxo\"><button class=\"svelte-1td2jxo\">Scale</button> <!></td></tr>"), jr = /* @__PURE__ */ Z("<h2 class=\"st\">Scale<!></h2> <!> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"scale\"><thead><tr><th>Pool</th><th>Desired</th><th>Open</th><th>Working</th><th>Idle</th><th>Cordoned</th><th>Min/max</th><th>Target</th><th></th></tr></thead><tbody></tbody></table></div>", 1), Mr = /* @__PURE__ */ Z("<tr><td class=\"mono c-id\"> </td><td class=\"c-st\"><span class=\"pill bad\">quarantined</span></td><td class=\"muted full\"> </td></tr>"), Nr = /* @__PURE__ */ Z("<tr><td class=\"mono c-id\"> </td><td class=\"c-st\"><span class=\"pill warn\">paused</span></td><td class=\"muted full\"> </td></tr>"), Pr = /* @__PURE__ */ Z("<tr><td class=\"faint\">All pools start sessions.</td></tr>"), Fr = /* @__PURE__ */ Z("<div class=\"faint\"> </div>"), Ir = /* @__PURE__ */ Z("<tr><td class=\"mono c-id\"> <!></td><td class=\"c-st\"><span> </span></td><td class=\"muted full\"> </td><td class=\"muted\" data-label=\"last fired\"> </td><td class=\"muted\" data-label=\"next cron\"> </td></tr>"), Lr = /* @__PURE__ */ Z("<tr><td class=\"faint\">No triggers.</td></tr>"), Rr = /* @__PURE__ */ Z("<!> <h2 class=\"st\">Pool status</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"pools\"><tbody><!><!><!></tbody></table></div> <h2 class=\"st\">Triggers</h2> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"triggers\"><thead><tr><th>Trigger</th><th>Holds</th><th>Why not / stalls</th><th>Last fired</th><th>Next cron</th></tr></thead><tbody></tbody></table></div>", 1), zr = /* @__PURE__ */ Z("<header class=\"vh\"><h1>Pools and triggers</h1></header> <div class=\"pad\"><!></div>", 1);
function Br(e, t) {
	Fe(t, !0);
	let n = xr(() => t.api, () => "api/pools"), r = /* @__PURE__ */ M("");
	async function i(e) {
		N(r, "");
		try {
			await t.api.post("api/pool/scale", e), n.refresh();
		} catch (e) {
			N(r, Sr(e), !0);
		}
	}
	function a(e) {
		let t = window.prompt("Target sessions for " + e.name + " (" + e.min + ".." + e.max + ")", String(e.target ?? e.desired));
		if (t === null || t.trim() === "") return;
		let n = Number(t);
		!Number.isInteger(n) || n < 0 ? N(r, "target must be a whole number") : i({
			pool: e.name,
			target: n
		});
	}
	var o = zr(), s = L(Ht(o), 2), c = F(s), l = (e) => {
		var t = wr(), r = I(t, !0);
		an(() => $(r, n.error)), Q(e, t);
	}, u = (e) => {
		let t = /* @__PURE__ */ $e(() => Object.entries(n.data.quarantined ?? {})), o = /* @__PURE__ */ $e(() => Object.entries(n.data.paused ?? {}));
		var s = Rr(), c = Ht(s), l = (e) => {
			var t = jr(), o = Ht(t), s = L(F(o)), c = (e) => {
				var t = Tr(), r = I(t);
				an(() => $(r, `(host max ${n.data.host_max ?? ""})`)), Q(e, t);
			};
			sr(s, (e) => {
				n.data.host_max && e(c);
			}), D(o);
			var l = L(o, 2), u = (e) => {
				var t = wr(), n = I(t, !0);
				an(() => $(n, X(r))), Q(e, t);
			};
			sr(l, (e) => {
				X(r) && e(u);
			});
			var d = L(l, 2), f = F(d), p = L(F(f));
			dr(p, 21, () => n.data.pools, (e) => e.name, (e, t) => {
				var n = Ar(), r = F(n), o = F(r, !0), s = L(o), c = (e) => {
					Q(e, Er());
				};
				sr(s, (e) => {
					X(t).drained && e(c);
				}), D(r);
				var l = L(r), u = I(l, !0), d = L(l), f = I(d, !0), p = L(d), m = I(p, !0), h = L(p), g = I(h, !0), _ = L(h), v = I(_, !0), y = L(_), b = I(y), x = L(y), S = F(x), ee = (e) => {
					var n = Dr(), r = I(n);
					an(() => $(r, `target ${X(t).target ?? ""}`)), Q(e, n);
				}, C = (e) => {
					Q(e, Or());
				};
				sr(S, (e) => {
					X(t).target == null ? e(C, -1) : e(ee);
				}), D(x);
				var te = L(x), ne = F(te), re = L(ne, 2), ie = (e) => {
					var n = kr();
					Bn("click", n, () => i({
						pool: X(t).name,
						auto: !0
					})), Q(e, n);
				};
				sr(re, (e) => {
					X(t).target != null && e(ie);
				}), D(te), D(n), an(() => {
					$(o, X(t).name), $(u, X(t).desired), $(f, X(t).open), $(m, X(t).working), $(g, X(t).idle), $(v, X(t).cordoned), $(b, `${X(t).min ?? ""}/${X(t).max ?? ""}`);
				}), Bn("click", ne, () => a(X(t))), Q(e, n);
			}), D(p), D(f), D(d), Q(e, t);
		};
		sr(c, (e) => {
			n.data.pools?.length && e(l);
		});
		var u = L(c, 4), d = F(u), f = F(d), p = F(f);
		dr(p, 17, () => X(t), ([e, t]) => e, (e, t) => {
			var n = /* @__PURE__ */ $e(() => h(X(t), 2));
			let r = () => X(n)[0], i = () => X(n)[1];
			var a = Mr(), o = F(a), s = I(o, !0), c = I(L(o, 2), !0);
			D(a), an(() => {
				$(s, r()), $(c, i());
			}), Q(e, a);
		});
		var m = L(p);
		dr(m, 17, () => X(o), ([e, t]) => e, (e, t) => {
			var n = /* @__PURE__ */ $e(() => h(X(t), 2));
			let r = () => X(n)[0], i = () => X(n)[1];
			var a = Nr(), o = F(a), s = I(o, !0), c = I(L(o, 2));
			D(a), an((e) => {
				$(s, r()), $(c, `until ${e ?? ""}`);
			}, [() => Cr(String(i()))]), Q(e, a);
		});
		var g = L(m), _ = (e) => {
			Q(e, Pr());
		};
		sr(g, (e) => {
			X(t).length === 0 && X(o).length === 0 && e(_);
		}), D(f), D(d), D(u);
		var v = L(u, 4), y = F(v), b = L(F(y));
		dr(b, 21, () => n.data.triggers, (e) => e.name, (e, t) => {
			var n = Ir(), r = F(n), i = F(r, !0), a = L(i), o = (e) => {
				var n = Fr(), r = I(n);
				an(() => $(r, `open ${X(t).open ?? ""}`)), Q(e, n);
			};
			sr(a, (e) => {
				X(t).open && e(o);
			}), D(r);
			var s = L(r), c = F(s), l = I(c, !0);
			D(s);
			var u = L(s), d = I(u, !0), f = L(u), p = I(f, !0), m = I(L(f), !0);
			D(n), an((e, n, r) => {
				$(i, X(t).name), yr(c, 1, `pill ${X(t).holds ? "accent" : ""}`), $(l, X(t).holds ? "fires next" : "waiting"), $(d, e), $(p, n), $(m, r);
			}, [
				() => X(t).stalls?.length ? X(t).stalls.join(", ") : X(t).why ?? "",
				() => Cr(X(t).last_fired),
				() => Cr(X(t).next_cron)
			]), Q(e, n);
		}, (e) => {
			Q(e, Lr());
		}), D(b), D(y), D(v), Q(e, s);
	};
	sr(c, (e) => {
		n.error ? e(l) : n.data && e(u, 1);
	}), D(s), Q(e, o), Ie();
}
// plugins/pools/index.js
Vn(["click"]), br("desk-pools-view", Br);
