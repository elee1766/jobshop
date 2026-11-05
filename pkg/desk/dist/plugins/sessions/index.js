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
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, C = 1 << 25, w = 65536, ee = 1 << 19, te = 1 << 20, T = 1 << 25, ne = 1 << 21, re = 1 << 22, ie = 1 << 23, ae = Symbol("$state"), oe = Symbol("component"), se = Symbol(""), ce = Symbol("attributes"), le = Symbol("class"), ue = Symbol("style"), de = Symbol("text"), fe = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), pe = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function me() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function he(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function E() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var D = !1;
function ge(e) {
	D = e;
}
var O;
function _e(t) {
	if (t === null) throw he(), e;
	return O = t;
}
function ve() {
	return _e(/* @__PURE__ */ R(O));
}
function ye(t) {
	if (D) {
		if (/* @__PURE__ */ R(O) !== null) throw he(), e;
		O = t;
	}
}
function be(e = 1) {
	if (D) {
		for (var t = e, n = O; t--;) n = /* @__PURE__ */ R(n);
		O = n;
	}
}
function xe(e = !0) {
	for (var t = 0, n = O;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ R(n);
		e && n.remove(), n = i;
	}
}
function Se(t) {
	if (!t || t.nodeType !== 8) throw he(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function k(e) {
	return e === this.v;
}
function Ce(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function A(e) {
	return !Ce(e, this.v);
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
var j = null;
function Pe(e) {
	j = e;
}
function Fe(e, t = !1, n) {
	j = {
		p: j,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: U,
		l: null
	};
}
function Ie(e) {
	var t = j, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) en(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, j = t.p, Le(e);
}
function Le(e = {}) {
	return s(e, oe, { value: !0 }), e;
}
function Re() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var ze = [];
function Be() {
	var e = ze;
	ze = [], h(e);
}
function Ve(e) {
	if (ze.length === 0 && !ut) {
		var t = ze;
		queueMicrotask(() => {
			t === ze && Be();
		});
	}
	ze.push(e);
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var He = ~(v | y | _);
function M(e, t) {
	e.f = e.f & He | t;
}
function Ue(e) {
	e.f & 512 || e.deps === null ? M(e, _) : M(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function We(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), M(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function Ge(e) {
	var t = H, n = U;
	Cn(null), wn(null);
	try {
		return e();
	} finally {
		Cn(t), wn(n);
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
	var s = U, c = qe(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				qt(e, s);
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
		Promise.all(n.map((e) => /* @__PURE__ */ Qe(e))).then(u).catch((e) => qt(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), Je());
	}) : f();
}
function qe() {
	var e = U, t = H, n = j, r = N;
	return function(i = !0) {
		wn(e), Cn(t), Pe(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function Je(e = !0) {
	wn(null), Cn(null), Pe(null), e && N?.deactivate();
}
function Ye() {
	var e = U, t = e.b, n = N, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function Xe(e) {
	var n = 2 | v;
	return U !== null && (U.f |= ee), {
		ctx: j,
		deps: null,
		effects: null,
		equals: k,
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
var Ze = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function Qe(e, n, r) {
	let i = U;
	i === null && we();
	var a = void 0, o = wt(t), s = !H, c = /* @__PURE__ */ new Set();
	return nn(() => {
		var t = U, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== fe && n.reject(e);
			}).finally(Je);
		} catch (e) {
			n.reject(e), Je();
		}
		var r = N;
		if (s) {
			if (t.f & 32768) var l = Ye();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(Ze);
			else for (let e of c.values()) e.reject(Ze);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== Ze && (r.activate(), t ? (o.f |= ie, Ot(o, t)) : (o.f & 8388608 && (o.f ^= ie), Ot(o, e)), r.deactivate());
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
	return En(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function et(e) {
	let t = /* @__PURE__ */ Xe(e);
	return t.equals = A, t;
}
function tt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) dn(t[n]);
	}
}
function nt(e) {
	var n, r = U, i = e.parent;
	if (!bn && i !== null && e.v !== t && i.f & 24576) return me(), e.v;
	wn(i);
	try {
		tt(e), n = Rn(e);
	} finally {
		wn(r);
	}
	return n;
}
function rt(e) {
	var t = nt(e);
	!e.equals(t) && (e.wv = Fn(), (!N?.is_fork || e.deps === null) && (N === null ? e.v = t : (N.capture(e, t, !0), st?.capture(e, t, !0)), e.deps === null)) ? M(e, _) : bn || (ct === null ? Ue(e) : (Zt() || N?.is_fork) && ct.set(e, t));
}
function it(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && Ge(() => {
		t.ac.abort(fe), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Vn(t, 0), ln(t));
}
function at(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Hn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var ot = null, N = null, st = null, ct = null, lt = null, ut = !1, dt = !1, ft = null, P = null, pt = 0, mt = 1, ht = class e {
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
			for (var r of n.d) M(r, v), t(r);
			for (r of n.m) M(r, y), t(r);
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
		for (let [e, t] of this.#s) M(e, t), this.schedule(e);
		this.apply();
		for (var t = ft = [], n = [], r = P = []; this.#a.length > 0;) {
			pt++ > 1e3 && (this.#b(), gt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw xt(e), this.#d() || this.discard(), t;
			}
		}
		if (N = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (ft = null, P = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) bt(e, t);
			r.length > 0 && N.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), st = this, vt(n), vt(t), st = null, this.#i?.resolve();
			var o = N;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (F.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : In(r) && (i & 16 && this.add_dirty_reaction(r, y), Hn(r));
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
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), M(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), N = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) We(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), ct?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		N = this;
	}
	deactivate() {
		N = null, ct = null;
	}
	flush() {
		try {
			dt = !0, N = this, this.#m();
		} finally {
			pt = 0, lt = null, ft = null, P = null, dt = !1, N = null, ct = null, F.clear();
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
		return (this.#i ??= g()).promise;
	}
	static ensure() {
		if (N === null) {
			let t = N = new e();
			!dt && Ve(() => {
				t.started || t.flush();
			});
		}
		return N;
	}
	apply() {
		ct = null;
	}
	schedule(e) {
		lt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
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
		qt(e, lt);
	}
}
var _t = null;
function vt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && In(r) && (_t = /* @__PURE__ */ new Set(), Hn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && pn(r), _t?.size > 0)) {
				F.clear();
				for (let e of _t) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) _t.has(n) && (_t.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Hn(n);
					}
				}
				_t.clear();
			}
		}
		_t = null;
	}
}
function yt(e) {
	N.schedule(e);
}
function bt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), M(e, _);
		for (var n = e.first; n !== null;) bt(n, t), n = n.next;
	}
}
function xt(e) {
	M(e, _);
	for (var t = e.first; t !== null;) xt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var St = /* @__PURE__ */ new Set(), F = /* @__PURE__ */ new Map(), Ct = !1;
function wt(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: k,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function I(e, t) {
	let n = wt(e, t);
	return En(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function Tt(e, t = !1, n = !0) {
	let r = wt(e);
	return t || (r.equals = A), r;
}
function L(e, t, n = !1) {
	return H !== null && (!Sn || H.f & 131072) && Re() && H.f & 4325394 && (Tn === null || !Tn.has(e)) && Me(), Ot(e, n ? Nt(t) : t, P);
}
var Et = null, Dt = 0;
function Ot(e, t, n = null) {
	if (!e.equals(t)) {
		bn ? F.set(e, t) : F.has(e) || F.set(e, e.v);
		var r = ht.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && nt(t), ct === null && Ue(t);
		}
		e.wv = Fn(), Et = null, Dt = 0, Mt(e, v, n), Et = null, Re() && U !== null && U.f & 1024 && !(U.f & 96) && (kn === null ? An([e]) : kn.push(e)), !r.is_fork && St.size > 0 && !Ct && kt();
	}
	return t;
}
function kt() {
	Ct = !1;
	for (let e of St) {
		e.f & 1024 && M(e, y);
		let t;
		try {
			t = In(e);
		} catch {
			t = !0;
		}
		t && Hn(e);
	}
	St.clear();
}
function At(e, t = 1) {
	var n = W(e), r = t === 1 ? n++ : n--;
	return L(e, n), r;
}
function jt(e) {
	L(e, e.v + 1);
}
function Mt(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Re(), a = r.length;
		if (Dt += a, Dt > 1e5 && Et === null && (Et = /* @__PURE__ */ new Set()), Et !== null) {
			if (Et.has(e)) return;
			Et.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== U) {
				var l = (c & v) === 0;
				if (l && M(s, t), c & 131072) St.add(s);
				else if (c & 2) {
					var u = s;
					ct?.delete(u), Mt(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && _t !== null && _t.add(d), n === null ? yt(d) : n.push(d);
				}
			}
		}
	}
}
function Nt(e) {
	if (typeof e != "object" || !e || ae in e || oe in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ I(0), s = null, l = Nn, p = (e) => {
		if (Nn === l) return e();
		var t = H, n = Nn;
		Cn(null), Pn(l);
		var r = e();
		return Cn(t), Pn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ I(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Ae();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ I(n.value, s);
				return i.set(t, e), e;
			}) : L(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ I(t, s));
					i.set(n, e), jt(o);
				}
			} else L(r, t), jt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === ae) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ I(Nt(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = W(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = W(a);
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
			if (n === ae) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (U !== null || H !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ I(a ? Nt(e[n]) : t, s)), i.set(n, r)), W(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ I(t, s)), i.set(f + "", m)) : L(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ I(void 0, s)), L(u, Nt(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => Nt(r));
				L(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && L(_, v + 1);
				}
				jt(o);
			}
			return !0;
		},
		ownKeys(e) {
			W(o);
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
var Pt, Ft, It, Lt;
function Rt() {
	if (Pt === void 0) {
		Pt = window, Ft = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		It = c(t, "firstChild").get, Lt = c(t, "nextSibling").get, p(e) && (e[le] = void 0, e[ce] = null, e[ue] = void 0, e.__e = void 0), p(n) && (n[de] = void 0);
	}
}
function zt(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Bt(e) {
	return It.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function R(e) {
	return Lt.call(e);
}
function z(e, t) {
	if (!D) return /* @__PURE__ */ Bt(e);
	var n = /* @__PURE__ */ Bt(O);
	if (n === null) n = O.appendChild(zt());
	else if (t && n.nodeType !== 3) {
		var r = zt();
		return n?.before(r), _e(r), r;
	}
	return t && Gt(n), _e(n), n;
}
function B(e, t = !1) {
	if (!D) {
		var n = /* @__PURE__ */ Bt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ R(n) : n;
	}
	if (t) {
		if (O?.nodeType !== 3) {
			var r = zt();
			return O?.before(r), _e(r), r;
		}
		Gt(O);
	}
	return O;
}
function Vt(e, t = !1) {
	if (!D) return /* @__PURE__ */ Bt(e);
	var n = z(e, t);
	return ye(e), n;
}
function V(e, t = 1, n = !1) {
	let r = D ? O : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ R(r);
	if (!D) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = zt();
			return r === null ? i?.after(a) : r.before(a), _e(a), a;
		}
		Gt(r);
	}
	return _e(r), r;
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
	var t = U;
	if (t === null) return H.f |= ie, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	qt(e, t);
}
function qt(t, n) {
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
	U === null && (H === null && Oe(e), De()), bn && Ee(e);
}
function Yt(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function Xt(e, t) {
	var n = U;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: j,
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
	N?.register_created_effect(r);
	var i = r;
	if (e & 4) ft === null ? ht.ensure().schedule(r) : ft.push(r);
	else if (t !== null) {
		try {
			Hn(r);
		} catch (e) {
			throw dn(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= w));
	}
	if (i !== null && (i.parent = n, n !== null && Yt(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function Zt() {
	return H !== null && !Sn;
}
function Qt(e) {
	let t = Xt(8, null);
	return M(t, _), t.teardown = e, t;
}
function $t(e) {
	Jt("$effect");
	var t = U.f;
	if (!H && t & 32 && j !== null && !j.i) {
		var n = j;
		(n.e ??= []).push(e);
	} else return en(e);
}
function en(e) {
	return Xt(4 | te, e);
}
function tn(e) {
	ht.ensure();
	let t = Xt(64 | ee, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? mn(t, () => {
			dn(t), n(void 0);
		}) : (dn(t), n(void 0));
	});
}
function nn(e) {
	return Xt(re | ee, e);
}
function rn(e, t = 0) {
	return Xt(8 | t, e);
}
function an(e, t = [], n = [], r = []) {
	Ke(r, t, n, (t) => {
		Xt(8, () => {
			e(...t.map(W));
		});
	});
}
function on(e, t = 0) {
	return Xt(16 | t, e);
}
function sn(e) {
	return Xt(32 | ee, e);
}
function cn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = bn, r = H;
		xn(!0), Cn(null);
		try {
			t.call(null);
		} catch (t) {
			qt(t, e.parent);
		} finally {
			xn(n), Cn(r);
		}
	}
}
function ln(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && Ge(() => {
			e.abort(fe);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : dn(n, t), n = r;
	}
}
function un(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || dn(t), t = n;
	}
}
function dn(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (fn(e.nodes.start, e.nodes.end), n = !0), e.f |= C, ln(e, t && !n), Vn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	cn(e), e.f ^= C, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && pn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function fn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ R(e);
		e.remove(), e = n;
	}
}
function pn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function mn(e, t, n = !0) {
	var r = [];
	e.f |= 256, hn(e, r, !0);
	var i = () => {
		n && dn(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function hn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				hn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function gn(e) {
	e.f &= -257, _n(e, !0);
}
function _n(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (M(e, v), ht.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			_n(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function vn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ R(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var yn = null, bn = !1;
function xn(e) {
	bn = e;
}
var H = null, Sn = !1;
function Cn(e) {
	H = e;
}
var U = null;
function wn(e) {
	U = e;
}
var Tn = null;
function En(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (Tn ??= /* @__PURE__ */ new Set()).add(e);
}
var Dn = null, On = 0, kn = null;
function An(e) {
	kn = e;
}
var jn = 1, Mn = 0, Nn = Mn;
function Pn(e) {
	Nn = e;
}
function Fn() {
	return ++jn;
}
function In(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (In(a) && rt(a), a.wv > e.wv) return !0;
		}
		t & 512 && ct === null && M(e, _);
	}
	return !1;
}
function Ln(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(Tn !== null && Tn.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Ln(a, t, !1) : t === a && (n ? M(a, v) : a.f & 1024 && M(a, y), yt(a));
	}
}
function Rn(e) {
	var t = Dn, n = On, r = kn, i = H, a = Tn, o = j, s = Sn, c = Nn, l = e.f;
	Dn = null, On = 0, kn = null, H = l & 96 ? null : e, Tn = null, Pe(e.ctx), Sn = !1, Nn = ++Mn, e.ac !== null && (Ge(() => {
		e.ac.abort(fe);
	}), e.ac = null);
	try {
		e.f |= ne;
		var u = e.fn, d = u();
		e.f |= S;
		var f = zn(e);
		if (Re() && kn !== null && !Sn && f !== null && !(e.f & 6146)) for (var p = 0; p < kn.length; p++) Ln(kn[p], e);
		if (i !== null && i !== e) {
			if (Mn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Mn;
			if (t !== null) for (let e of t) e.rv = Mn;
			kn !== null && (r === null ? r = kn : r.push(...kn));
		}
		return e.f & 8388608 && (e.f ^= ie), d;
	} catch (t) {
		return zn(e), Kt(t);
	} finally {
		e.f ^= ne, Dn = t, On = n, kn = r, H = i, Tn = a, Pe(o), Sn = s, Nn = c;
	}
}
function zn(e) {
	var t = e.deps, n = N?.is_fork;
	if (Dn !== null) {
		var r;
		if (n || Vn(e, On), t !== null && On > 0) for (t.length = On + Dn.length, r = 0; r < Dn.length; r++) t[On + r] = Dn[r];
		else e.deps = t = Dn;
		if (Zt() && e.f & 512) for (r = On; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && On < t.length && (Vn(e, On), t.length = On);
	return t;
}
function Bn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (Dn === null || !a.call(Dn, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ue(c), c.ac !== null && Ge(() => {
			c.ac.abort(fe), c.ac = null, M(c, v);
		}), it(c), Vn(c, 0);
	}
}
function Vn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Bn(e, n[r]);
}
function Hn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		M(e, _);
		var n = U;
		U = e;
		try {
			t & 16777232 ? un(e) : ln(e), cn(e);
			var r = Rn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = jn;
		} finally {
			U = n;
		}
	}
}
function W(e) {
	var t = !!(e.f & 2);
	if (yn?.add(e), H !== null && !Sn && !(U !== null && U.f & 16384) && (Tn === null || !Tn.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < Mn && (e.rv = Mn, Dn === null && n !== null && n[On] === e ? On++ : Dn === null ? Dn = [e] : Dn.push(e));
		else {
			H.deps ??= [], a.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : a.call(r, H) || r.push(H);
		}
	}
	if (bn && F.has(e)) return F.get(e);
	if (t) {
		var i = e;
		if (bn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Wn(i)) && (o = nt(i)), F.set(i, o), o;
		}
		var s = !(i.f & 512) && !Sn && H !== null && !!(H.f & 512), c = (i.f & S) === 0;
		In(i) && (s && (i.f |= 512), rt(i)), s && !c && (at(i), Un(i));
	}
	if (ct?.has(e)) return ct.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Un(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (at(n), Un(n));
	}
}
function Wn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (F.has(t) || t.f & 2 && Wn(t)) return !0;
	return !1;
}
function Gn(e) {
	var t = Sn;
	try {
		return Sn = !0, e();
	} finally {
		Sn = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var Kn = Symbol("events"), qn = /* @__PURE__ */ new Set(), Jn = /* @__PURE__ */ new Set();
function Yn(e, t, n) {
	(t[Kn] ??= {})[e] = n;
}
function Xn(e) {
	for (var t = 0; t < e.length; t++) qn.add(e[t]);
	for (var n of Jn) n(e);
}
var Zn = null, Qn = !1;
function $n(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	Zn = e, Qn || (Qn = !0, setTimeout(() => {
		Qn = !1, Zn = null;
	}));
	var o = 0, c = Zn === e && e[Kn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[Kn] = t;
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
		var d = H, f = U;
		Cn(null), wn(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[Kn]?.[r];
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
			e[Kn] = t, delete e.currentTarget, Cn(d), wn(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var er = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function tr(e) {
	return er?.createHTML(e) ?? e;
}
function nr(e) {
	var t = Wt("template");
	return t.innerHTML = tr(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function rr(e, t) {
	var n = U;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function ir(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (D) return rr(O, null), O;
		i === void 0 && (i = nr(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Bt(i)));
		var t = r || Ft ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Bt(t), s = t.lastChild;
			rr(o, s);
		} else rr(t, t);
		return t;
	};
}
function ar() {
	if (D) return rr(O, null), O;
	var e = document.createDocumentFragment(), t = document.createComment(""), n = zt();
	return e.append(t, n), rr(t, n), e;
}
function or(e, t) {
	if (D) {
		var n = U;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = O), ve();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var sr = ["touchstart", "touchmove"];
function cr(e) {
	return sr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function lr(e) {
	let t = 0, n = wt(0), r;
	return () => {
		Zt() && (W(n), rn(() => (t === 0 && (r = Gn(() => e(() => jt(n)))), t += 1, () => {
			Ve(() => {
				--t, t === 0 && (r?.(), r = void 0, jt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var ur = w | ee;
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
	#m = lr(() => (this.#p = wt(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = U;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = U.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = on(() => {
			if (D) {
				let e = Se(this.#t);
				ve();
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
			this.#a = sn(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ve(r), t && (this.#s = sn(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? E() : (t = !0, n && Ne(), this.#s !== null && mn(this.#s, () => {
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
					qt(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | C)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = sn(() => e(this.#e)), Ve(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = zt(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return sn(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							qt(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(N)) : this.#u === 0 && (this.#e.before(e), this.#c = null, mn(this.#o, () => {
					this.#o = null;
				}), this.#x(N));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = sn(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				vn(this.#a, e);
				let t = this.#n.pending;
				this.#o = sn(() => t(this.#e));
			} else this.#x(N);
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
		var t = U, n = H, r = j;
		wn(this.#i), Cn(this.#i), Pe(this.#i.ctx);
		try {
			return ht.ensure(), e();
		} finally {
			wn(t), Cn(n), Pe(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && mn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ve(() => {
			this.#d = !1, this.#p && Ot(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), W(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		N?.is_fork ? (this.#a && N.skip_effect(this.#a), this.#o && N.skip_effect(this.#o), this.#s && N.skip_effect(this.#s), N.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (dn(this.#a), null), this.#o &&= (dn(this.#o), null), this.#s &&= (dn(this.#s), null), D && (_e(this.#t), be(), _e(xe()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return sn(() => {
						var r = U;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return qt(e, this.#i.parent), null;
				}
			}));
		};
		Ve(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					qt(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => qt(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function pr(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[de] ??= e.nodeValue) && (e[de] = n, e.nodeValue = `${n}`);
}
function mr(e, t) {
	return gr(e, t);
}
var hr = /* @__PURE__ */ new Map();
function gr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Rt();
	var u = void 0, d = tn(() => {
		var c = r ?? n.appendChild(zt());
		dr(c, { pending: () => {} }, (n) => {
			Fe({});
			var r = j;
			if (s && (r.c = s), a && (i.$$events = a), D && rr(n, null), u = t(n, i) || Le(), D && (U.nodes.end = O, O === null || O.nodeType !== 8 || O.data !== "]")) throw he(), e;
			Ie();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = cr(r);
					for (let e of [n, document]) {
						var a = hr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), hr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, $n, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(qn)), Jn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = hr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, $n), t.delete(e), t.size === 0 && hr.delete(r)) : t.set(e, i);
			}
			Jn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return _r.set(u, d), u;
}
var _r = /* @__PURE__ */ new WeakMap();
function vr(e, t) {
	let n = _r.get(e);
	return n ? (_r.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var yr = class {
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
			if (n) gn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (gn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (dn(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						vn(r, t), t.append(zt()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else dn(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), mn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (dn(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = N, r = Ut();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = zt();
				i.append(a), this.#n.set(e, {
					effect: sn(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, sn(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else D && (this.anchor = O), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function br(e, t, n = !1) {
	var r;
	D && (r = O, ve());
	var i = new yr(e), a = n ? w : 0;
	function o(e, t) {
		if (D) {
			var n = Se(r);
			if (e !== parseInt(n.substring(1))) {
				var a = xe();
				_e(a), i.anchor = a, ge(!1), i.ensure(e, t), ge(!0);
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
function xr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		mn(n, () => {
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
			Ht(d), d.append(u), e.items.clear();
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
		r?.has(a) ? (a.f |= T, vn(a, document.createDocumentFragment())) : dn(t[i], n);
	}
}
var Cr;
function wr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = D ? _e(/* @__PURE__ */ Bt(u)) : u.appendChild(zt());
	}
	D && ve();
	var d = null, f = /* @__PURE__ */ et(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = W(f);
			_.fallback = d, Er(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= T, Or(d, null, c)) : gn(d) : mn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: on(() => {
			var e = W(f), r = e.length;
			let o = !1;
			D && Se(c) === "[!" != (r === 0) && (c = xe(), _e(c), ge(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = N, v = Ut(), y = 0; y < r; y += 1) {
				D && O.nodeType === 8 && O.data === "]" && (c = O, o = !0, ge(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && Ot(S.v, b), S.i && Ot(S.i, y), v && _.unskip_effect(S.e)) : (S = Dr(l, m ? c : Cr ??= zt(), b, x, y, a, t, n), m || (S.e.f |= T), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = sn(() => s(c)) : (d = sn(() => s(Cr ??= zt())), d.f |= T)), r > u.size && Te("", "", ""), D && r > 0 && _e(xe()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && ge(!0), W(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, D && (c = O);
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
		if (_.f & 8192 && (gn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= T, _ === l) Or(_, null, n);
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
		var w = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || w.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && w.push(l), l = Tr(l.next);
		var ee = w.length;
		if (ee > 0) {
			var te = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.measure();
				for (v = 0; v < ee; v += 1) w[v].nodes?.a?.fix();
			}
			xr(e, w, te);
		}
	}
	a && Ve(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Dr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? wt(n) : /* @__PURE__ */ Tt(n, !1, !1) : null, l = o & 2 ? wt(i) : null;
	return {
		v: c,
		i: l,
		e: sn(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Or(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ R(r);
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
	var o = e[le];
	if (D || o !== n || o === void 0) {
		var s = jr(n, r, a);
		D && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[le] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Nr = Symbol("is custom element"), Pr = Symbol("is html"), Fr = pe ? "link" : "LINK";
function Ir(e, t, n, r) {
	var i = Lr(e);
	if (!(D && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Fr))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[se] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && zr(e).has(t) ? e[t] = n : (!D || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Lr(e) {
	return e[ce] ??= {
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
// kit/plugin.svelte.js
function Br(e, t) {
	if (customElements.get(e)) return;
	let n = new URL(
		/* @vite-ignore */
		"./style.css",
		import.meta.url
	).href;
	customElements.define(e, class extends HTMLElement {
		#e = /* @__PURE__ */ I(Nt({
			api: null,
			context: {},
			config: {}
		}));
		get props() {
			return W(this.#e);
		}
		set props(e) {
			L(this.#e, e, !0);
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
			this.style.display = "contents", this.app ||= mr(t, {
				target: this,
				props: this.props
			});
		}
		disconnectedCallback() {
			this.app && vr(this.app), this.app = null;
		}
	});
}
function Vr(e, t) {
	let n = Nt({
		data: null,
		error: "",
		loading: !0,
		refresh: () => {
			At(r);
		}
	}), r = /* @__PURE__ */ I(0);
	$t(() => {
		let t = e();
		if (!t) return;
		let n = 0, i = t.subscribe(() => {
			n ||= window.setTimeout(() => {
				n = 0, At(r);
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
		W(r), n && c && (c === o && i ? a = !0 : (o = c, s(n, c)));
	}), n;
}
function Hr(e) {
	let t = Nt({});
	return $t(() => {
		let n = e();
		if (!n) return;
		let r = (e) => {
			for (let n of Object.keys(t)) n in e || delete t[n];
			Object.assign(t, e);
		};
		return r(n.params()), n.onParams(r);
	}), t;
}
var Ur = (e) => String(e?.message || e);
// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/marked/lib/marked.esm.js
function Wr() {
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
var Gr = Wr();
function Kr(e) {
	Gr = e;
}
var qr = { exec: () => null };
function Jr(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function G(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(K.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var Yr = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), K = {
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
	nextBulletRegex: Jr((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: Jr((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: Jr((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: Jr((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: Jr((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: Jr((e) => RegExp(`^ {0,${e}}(?:</?(?:${ui})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: Jr((e) => RegExp(`^ {0,${e}}>`))
}, Xr = /^(?:[ \t]*(?:\n|$))+/, Zr = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Qr = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, $r = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, ei = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, ti = / {0,3}(?:[*+-]|\d{1,9}[.)])/, ni = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, ri = G(ni).replace(/bull/g, ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), ii = G(ni).replace(/bull/g, ti).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), ai = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, oi = /^[^\n]+/, si = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, ci = G(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", si).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), li = G(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, ti).getRegex(), ui = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", di = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, fi = G("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", di).replace("tag", ui).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), pi = (e) => G(ai).replace("hr", $r).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", ui).getRegex(), mi = pi(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), hi = pi(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), gi = {
	blockquote: G(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", hi).getRegex(),
	code: Zr,
	def: ci,
	fences: Qr,
	heading: ei,
	hr: $r,
	html: fi,
	lheading: ri,
	list: li,
	newline: Xr,
	paragraph: mi,
	table: qr,
	text: oi
}, _i = G("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", $r).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", ui).getRegex(), vi = {
	...gi,
	lheading: ii,
	table: _i,
	paragraph: G(ai).replace("hr", $r).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", _i).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", ui).getRegex()
}, yi = {
	...gi,
	html: G("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", di).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: qr,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: G(ai).replace("hr", $r).replace("heading", " *#{1,6} *[^\n]").replace("lheading", ri).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, bi = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, xi = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Si = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Ci = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, wi = /[\p{P}\p{S}]/u, Ti = /[\s\p{P}\p{S}]/u, Ei = /[^\s\p{P}\p{S}]/u, Di = G(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Ti).getRegex(), Oi = /[\p{Pi}\p{Ps}"']/u, ki = /(?!~)[\p{P}\p{S}]/u, Ai = /(?!~)[\s\p{P}\p{S}]/u, ji = /(?:[^\s\p{P}\p{S}]|~)/u, Mi = G(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", Yr ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), Ni = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, Pi = G(Ni, "u").replace(/punct/g, wi).getRegex(), Fi = G(Ni, "u").replace(/punct/g, ki).getRegex(), Ii = G(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, Oi).replace(/punct/g, wi).getRegex(), Li = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", Ri = G(Li, "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), zi = G(Li, "gu").replace(/notPunctSpace/g, ji).replace(/punctSpace/g, Ai).replace(/punct/g, ki).getRegex(), Bi = G("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Vi = G("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Hi = G("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Ui = G(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, wi).getRegex(), Wi = G("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Ei).replace(/punctSpace/g, Ti).replace(/punct/g, wi).getRegex(), Gi = G(/\\(punct)/, "gu").replace(/punct/g, wi).getRegex(), Ki = G(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, wi).getRegex(), qi = G(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), Ji = G(di).replace("(?:-->|$)", "-->").getRegex(), Yi = G("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", Ji).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Xi = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, Zi = G(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", Xi).getRegex(), Qi = G(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", Zi).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), $i = G(/^!?\[(label)\]\[(ref)\]/).replace("label", Zi).replace("ref", si).getRegex(), ea = G(/^!?\[(ref)\](?:\[\])?/).replace("ref", si).getRegex(), ta = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, na = G(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", Xi).getRegex(), ra = G("reflink|nolink(?!\\()", "g").replace("reflink", G(/^!?\[(label)\]\[(ref)\]/).replace("label", na).replace("ref", ta).getRegex()).replace("nolink", G(/^!?\[(ref)\](?:\[\])?/).replace("ref", ta).getRegex()).getRegex(), ia = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, aa = G(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), oa = {
	_backpedal: qr,
	anyPunctuation: Gi,
	anyPunctuationOutsideCode: Ki,
	autolink: qi,
	blockSkip: Mi,
	br: Si,
	code: xi,
	del: qr,
	delLDelim: qr,
	delRDelim: qr,
	emStrongLDelim: Pi,
	emStrongRDelimAst: Ri,
	emStrongRDelimUnd: Vi,
	escape: bi,
	link: Qi,
	nolink: ea,
	punctuation: Di,
	reflink: $i,
	reflinkSearch: ra,
	tag: Yi,
	text: Ci,
	url: qr
}, sa = {
	...oa,
	emStrongLDelim: Ii,
	emStrongRDelimAst: Bi,
	emStrongRDelimUnd: Hi,
	link: G(/^!?\[(label)\]\((.*?)\)/).replace("label", Zi).getRegex(),
	reflink: G(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Zi).getRegex()
}, ca = {
	...oa,
	emStrongRDelimAst: zi,
	emStrongLDelim: Fi,
	delLDelim: Ui,
	delRDelim: Wi,
	url: G(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", aa).replace("protocol", ia).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: G(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", ia).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", Yr ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, la = {
	...ca,
	br: G(Si).replace("{2,}", "*").getRegex(),
	text: G(ca.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, ua = {
	normal: gi,
	gfm: vi,
	pedantic: yi
}, da = {
	normal: oa,
	gfm: ca,
	breaks: la,
	pedantic: sa
}, fa = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, pa = (e) => fa[e];
function ma(e, t) {
	if (t) {
		if (K.escapeTest.test(e)) return e.replace(K.escapeReplace, pa);
	} else if (K.escapeTestNoEncode.test(e)) return e.replace(K.escapeReplaceNoEncode, pa);
	return e;
}
function ha(e) {
	return K.numericCharacterReference.test(e) ? e.replace(K.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function ga(e) {
	try {
		e = encodeURI(e).replace(K.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function _a(e, t) {
	let n = e.replace(K.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(K.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(K.slashPipe, "|");
	return n;
}
function va(e, t, n) {
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
function ya(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && K.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function ba(e) {
	return e.trim().replace(K.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function xa(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Sa = /* @__PURE__ */ new WeakMap();
function Ca(e) {
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
function wa(e, t) {
	let n = Ca(t), r = Sa.get(e);
	r ? r.push(n) : Sa.set(e, [n]);
}
function Ta(e) {
	let t = Sa.get(e);
	t && (t.pop(), t.length === 0 && Sa.delete(e));
}
function Ea(e, t, n) {
	let r = Sa.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : Ca(t).lastPossibleEmStrongCloserIndex[n];
}
function Da(e, t) {
	let n = Sa.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : Ca(t).lastPossibleDelCloserIndex;
}
function Oa(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function ka(e, t, n, r, i) {
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
function Aa(e, t, n) {
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
function ja(e, t, n, r) {
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
var Ma = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || Gr;
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
			let e = this.options.pedantic ? t[0] : ya(t[0]);
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
			let e = t[0], n = Aa(e, t[3] || "", this.rules);
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
				let t = va(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: va(t[0], "\n"),
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
			raw: va(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = va(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + Oa(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
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
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? Oa(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => Oa(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => Oa(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
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
			let e = ya(t[0]);
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
			if (!this.rules.other.startAngleBracket.test(t[2]) && xa(t[2], "()") !== -1) return;
			let e = ba(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: va(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = _a(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: va(t[0], "\n"),
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
			for (let e of i) a.rows.push(_a(e, a.header.length).map((e, t) => ({
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
				raw: va(t[0], "\n"),
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
			if (!this.options.pedantic && ja(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = va(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = xa(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = va(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), ka(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && ja(e, n[1], r, this.rules)) return;
			let i = t[ba(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return ka(n, i, n[0], this.lexer, this.rules);
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
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Ea(this, t, i), y = t.length - e.length + n;
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Da(this, t), l = t.length - e.length + n;
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
				text: e ? t[0] : ha(t[0]),
				escaped: e
			};
		}
	}
}, Na = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || Gr, this.options.tokenizer = this.options.tokenizer || new Ma(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: K,
			block: ua.normal,
			inline: da.normal
		};
		this.options.pedantic ? (t.block = ua.pedantic, t.inline = da.pedantic) : this.options.gfm && (t.block = ua.gfm, t.inline = this.options.breaks ? da.breaks : da.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: ua,
			inline: da
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(K.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(K.tabCharGlobal, "    ").replace(K.spaceLine, ""));
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
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, ba(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
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
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== Ma.prototype.emStrong || this.tokenizer.del !== Ma.prototype.del ? this.#n(e) : e;
		wa(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Ta(this.tokenizer);
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
				if (!Object.hasOwn(this.tokens.links, ba(t.slice(r + 1, -1)))) return t;
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
}, Pa = class {
	options;
	parser;
	constructor(e) {
		this.options = e || Gr;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(K.notSpaceStart)?.[0], i = e ? e.replace(K.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + ma(r) + "\">" + (n ? i : ma(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : ma(i, !0)) + "</code></pre>\n";
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
		return `<code>${ma(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? ma(n, !0) : this.parser.parseInline(r), o = ga(e);
		if (o === null) return a;
		e = ma(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + ma(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = ga(e);
		if (i === null) return ma(n);
		e = i;
		let a = `<img src="${ma(e)}" alt="${ma(n)}"`;
		return t && (a += ` title="${ma(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : ma(e.text);
	}
}, Fa = class {
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
}, Ia = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || Gr, this.options.renderer = this.options.renderer || new Pa(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Fa(), this.textRenderer.parser = this;
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
}, La = class {
	options;
	block;
	constructor(e) {
		this.options = e || Gr;
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
		return e ? Na.lex : Na.lexInline;
	}
	provideParser(e = this.block) {
		return e ? Ia.parse : Ia.parseInline;
	}
};
function Ra(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var za = new class {
	defaults = Wr();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = Ia;
	Renderer = Pa;
	TextRenderer = Fa;
	Lexer = Na;
	Tokenizer = Ma;
	Hooks = La;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (Ra(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) Ra(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) Ra(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				Ra(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					Ra(n, this.walkTokens(i, t));
				}) : e.tokens && Ra(n, this.walkTokens(e.tokens, t));
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
				let t = this.defaults.renderer || new Pa(this.defaults);
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
				let t = this.defaults.tokenizer || new Ma(this.defaults);
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
				let t = this.defaults.hooks || new La();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = La.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && La.passThroughHooksRespectAsync.has(n)) return (async () => {
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
					return Ra(n, r.call(this, e)), t && Ra(n, t.call(this, e)), n;
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
		return Na.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return Ia.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? Na.lex : Na.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? Ia.parse : Ia.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? Na.lex : Na.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? Ia.parse : Ia.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + ma(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function q(e, t) {
	return za.parse(e, t);
}
q.options = q.setOptions = function(e) {
	return za.setOptions(e), q.defaults = za.defaults, Kr(q.defaults), q;
}, q.getDefaults = Wr, q.defaults = Gr;
function Ba(...e) {
	return za.use(...e), q.defaults = za.defaults, Kr(q.defaults), q;
}
q.use = Ba, q.walkTokens = function(e, t) {
	return za.walkTokens(e, t);
}, q.parseInline = za.parseInline, q.Parser = Ia, q.parser = Ia.parse, q.Renderer = Pa, q.TextRenderer = Fa, q.Lexer = Na, q.lexer = Na.lex, q.Tokenizer = Ma, q.Hooks = La, q.parse = q, q.options, q.setOptions, q.walkTokens, q.parseInline, Ia.parse, Na.lex;
// node_modules/dompurify/dist/purify.es.mjs
function Va(e, t) {
	this.v = e, this.k = t;
}
function Ha(e, t) {
	(t == null || t > e.length) && (t = e.length);
	for (var n = 0, r = Array(t); n < t; n++) r[n] = e[n];
	return r;
}
function Ua(e) {
	if (Array.isArray(e)) return e;
}
function Wa(e, t) {
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
function Ga() {
	throw TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function Ka(e, t) {
	return Ua(e) || Wa(e, t) || qa(e, t) || Ga();
}
function qa(e, t) {
	if (e) {
		if (typeof e == "string") return Ha(e, t);
		var n = {}.toString.call(e).slice(8, -1);
		return n === "Object" && e.constructor && (n = e.constructor.name), n === "Map" || n === "Set" ? Array.from(e) : n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n) ? Ha(e, t) : void 0;
	}
}
function Ja(e) {
	var t, n;
	function r(t, n) {
		try {
			var a = e[t](n), o = a.value, s = o instanceof Va;
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
Ja.prototype[typeof Symbol == "function" && Symbol.asyncIterator || "@@asyncIterator"] = function() {
	return this;
}, Ja.prototype.next = function(e) {
	return this._invoke("next", e);
}, Ja.prototype.throw = function(e) {
	return this._invoke("throw", e);
}, Ja.prototype.return = function(e) {
	return this._invoke("return", e);
};
var Ya = Object.entries, Xa = Object.setPrototypeOf, Za = Object.isFrozen, Qa = Object.getPrototypeOf, $a = Object.getOwnPropertyDescriptor, J = Object.freeze, Y = Object.seal, eo = Object.create, to = typeof Reflect < "u" && Reflect, no = to.apply, ro = to.construct;
J ||= function(e) {
	return e;
}, Y ||= function(e) {
	return e;
}, no ||= function(e, t) {
	var n = [...arguments].slice(2);
	return e.apply(t, n);
}, ro ||= function(e) {
	return new e(...[...arguments].slice(1));
};
var io = Z(Array.prototype.forEach);
Array.prototype.indexOf;
var ao = Z(Array.prototype.lastIndexOf), oo = Z(Array.prototype.pop), so = Z(Array.prototype.push), co = Z(Array.prototype.slice), lo = Z(Array.prototype.splice), uo = Array.isArray, fo = Oo(Set), po = Z(Set.prototype.add), mo = Z(Set.prototype.has), ho = Z(String.prototype.toLowerCase), go = Z(String.prototype.toString), _o = Z(String.prototype.match), vo = Z(String.prototype.replace), yo = Z(String.prototype.indexOf), bo = Z(String.prototype.trim), xo = Z(Number.prototype.toString), So = Z(Boolean.prototype.toString), Co = typeof BigInt > "u" ? null : Z(BigInt.prototype.toString), wo = typeof Symbol > "u" ? null : Z(Symbol.prototype.toString), To = Z(Object.prototype.hasOwnProperty), Eo = Z(Object.prototype.toString), X = Z(RegExp.prototype.test), Do = Oo(TypeError);
function Z(e) {
	return function(t) {
		t instanceof RegExp && (t.lastIndex = 0);
		var n = [...arguments].slice(1);
		return no(e, t, n);
	};
}
function Oo(e) {
	return function() {
		return ro(e, [...arguments]);
	};
}
function Q(e, t) {
	let n = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : ho;
	if (Xa && Xa(e, null), !uo(t)) return e;
	let r = t.length;
	for (; r--;) {
		let i = t[r];
		if (typeof i == "string") {
			let e = n(i);
			e !== i && (Za(t) || (t[r] = e), i = e);
		}
		e[i] = !0;
	}
	return e;
}
function ko(e) {
	for (let t = 0; t < e.length; t++) To(e, t) || (e[t] = null);
	return e;
}
function Ao(e) {
	let t = eo(null);
	for (let r of Ya(e)) {
		var n = Ka(r, 2);
		let i = n[0], a = n[1];
		To(e, i) && (t[i] = uo(a) ? ko(a) : a && typeof a == "object" && a.constructor === Object ? Ao(a) : a);
	}
	return t;
}
function jo(e) {
	switch (typeof e) {
		case "string": return e;
		case "number": return xo(e);
		case "boolean": return So(e);
		case "bigint": return Co ? Co(e) : "0";
		case "symbol": return wo ? wo(e) : "Symbol()";
		case "undefined": return Eo(e);
		case "function":
		case "object": {
			if (e === null) return Eo(e);
			let t = e, n = Mo(t, "toString");
			if (typeof n == "function") {
				let e = n(t);
				return typeof e == "string" ? e : Eo(e);
			}
			return Eo(e);
		}
		default: return Eo(e);
	}
}
function Mo(e, t) {
	for (; e !== null;) {
		let n = $a(e, t);
		if (n) {
			if (n.get) return Z(n.get);
			if (typeof n.value == "function") return Z(n.value);
		}
		e = Qa(e);
	}
	function n() {
		return null;
	}
	return n;
}
function No(e, t) {
	for (; e !== null;) {
		let n = $a(e, t);
		if (n) return n.get ? Z(n.get) : null;
		e = Qa(e);
	}
	return null;
}
function Po(e) {
	try {
		return X(e, ""), !0;
	} catch {
		return !1;
	}
}
var Fo = J(/* @__PURE__ */ "a.abbr.acronym.address.area.article.aside.audio.b.bdi.bdo.big.blink.blockquote.body.br.button.canvas.caption.center.cite.code.col.colgroup.content.data.datalist.dd.decorator.del.details.dfn.dialog.dir.div.dl.dt.element.em.fieldset.figcaption.figure.font.footer.form.h1.h2.h3.h4.h5.h6.head.header.hgroup.hr.html.i.img.input.ins.kbd.label.legend.li.main.map.mark.marquee.menu.menuitem.meter.nav.nobr.ol.optgroup.option.output.p.picture.pre.progress.q.rp.rt.ruby.s.samp.search.section.select.shadow.slot.small.source.spacer.span.strike.strong.style.sub.summary.sup.table.tbody.td.template.textarea.tfoot.th.thead.time.tr.track.tt.u.ul.var.video.wbr".split(".")), Io = J(/* @__PURE__ */ "svg.a.altglyph.altglyphdef.altglyphitem.animatecolor.animatemotion.animatetransform.circle.clippath.defs.desc.ellipse.enterkeyhint.exportparts.filter.font.g.glyph.glyphref.hkern.image.inputmode.line.lineargradient.marker.mask.metadata.mpath.part.path.pattern.polygon.polyline.radialgradient.rect.stop.style.switch.symbol.text.textpath.title.tref.tspan.view.vkern".split(".")), Lo = J([
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
]), Ro = J([
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
]), zo = J(/* @__PURE__ */ "math.menclose.merror.mfenced.mfrac.mglyph.mi.mlabeledtr.mmultiscripts.mn.mo.mover.mpadded.mphantom.mroot.mrow.ms.mspace.msqrt.mstyle.msub.msup.msubsup.mtable.mtd.mtext.mtr.munder.munderover.mprescripts".split(".")), Bo = J([
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
]), Vo = J(["#text"]), Ho = J(/* @__PURE__ */ "accept.action.align.alt.autocapitalize.autocomplete.autopictureinpicture.autoplay.background.bgcolor.border.capture.cellpadding.cellspacing.checked.cite.class.clear.color.cols.colspan.command.commandfor.controls.controlslist.coords.crossorigin.datetime.decoding.default.dir.disabled.disablepictureinpicture.disableremoteplayback.download.draggable.enctype.enterkeyhint.exportparts.face.for.headers.height.hidden.high.href.hreflang.id.inert.inputmode.integrity.ismap.kind.label.lang.list.loading.loop.low.max.maxlength.media.method.min.minlength.multiple.muted.name.nonce.noshade.novalidate.nowrap.open.optimum.part.pattern.placeholder.playsinline.popover.popovertarget.popovertargetaction.poster.preload.pubdate.radiogroup.readonly.rel.required.rev.reversed.role.rows.rowspan.spellcheck.scope.selected.shape.size.sizes.slot.span.srclang.start.src.srcset.step.style.summary.tabindex.title.translate.type.usemap.valign.value.width.wrap.xmlns".split(".")), Uo = J(/* @__PURE__ */ "accent-height.accumulate.additive.alignment-baseline.amplitude.ascent.attributename.attributetype.azimuth.basefrequency.baseline-shift.begin.bias.by.class.clip.clippathunits.clip-path.clip-rule.color.color-interpolation.color-interpolation-filters.color-profile.color-rendering.cx.cy.d.dx.dy.diffuseconstant.direction.display.divisor.dominant-baseline.dur.edgemode.elevation.end.exponent.fill.fill-opacity.fill-rule.filter.filterunits.flood-color.flood-opacity.font-family.font-size.font-size-adjust.font-stretch.font-style.font-variant.font-weight.fx.fy.g1.g2.glyph-name.glyphref.gradientunits.gradienttransform.height.href.id.image-rendering.in.in2.intercept.k.k1.k2.k3.k4.kerning.keypoints.keysplines.keytimes.lang.lengthadjust.letter-spacing.kernelmatrix.kernelunitlength.lighting-color.local.marker-end.marker-mid.marker-start.markerheight.markerunits.markerwidth.maskcontentunits.maskunits.max.mask.mask-type.media.method.mode.min.name.numoctaves.offset.operator.opacity.order.orient.orientation.origin.overflow.paint-order.path.pathlength.patterncontentunits.patterntransform.patternunits.pointer-events.points.preservealpha.preserveaspectratio.primitiveunits.r.rx.ry.radius.refx.refy.repeatcount.repeatdur.restart.result.rotate.scale.seed.shape-rendering.slope.specularconstant.specularexponent.spreadmethod.startoffset.stddeviation.stitchtiles.stop-color.stop-opacity.stroke-dasharray.stroke-dashoffset.stroke-linecap.stroke-linejoin.stroke-miterlimit.stroke-opacity.stroke.stroke-width.style.surfacescale.systemlanguage.tabindex.tablevalues.targetx.targety.transform.transform-origin.text-anchor.text-decoration.text-orientation.text-rendering.textlength.type.u1.u2.unicode.values.vector-effect.viewbox.visibility.version.vert-adv-y.vert-origin-x.vert-origin-y.width.word-spacing.wrap.writing-mode.xchannelselector.ychannelselector.x.x1.x2.xmlns.y.y1.y2.z.zoomandpan".split(".")), Wo = J(/* @__PURE__ */ "accent.accentunder.align.bevelled.close.columnalign.columnlines.columnspacing.columnspan.denomalign.depth.dir.display.displaystyle.encoding.fence.frame.height.href.id.largeop.length.linethickness.lquote.lspace.mathbackground.mathcolor.mathsize.mathvariant.maxsize.minsize.movablelimits.notation.numalign.open.rowalign.rowlines.rowspacing.rowspan.rspace.rquote.scriptlevel.scriptminsize.scriptsizemultiplier.selection.separator.separators.stretchy.subscriptshift.supscriptshift.symmetric.voffset.width.xmlns".split(".")), Go = J([
	"xlink:href",
	"xml:id",
	"xlink:title",
	"xml:space",
	"xmlns:xlink"
]), Ko = Y(/{{[\w\W]*|^[\w\W]*}}/g), qo = Y(/<%[\w\W]*|^[\w\W]*%>/g), Jo = Y(/\${[\w\W]*/g), Yo = Y(/^data-[\-\w.\u00B7-\uFFFF]+$/), Xo = Y(/^aria-[\-\w]+$/), Zo = Y(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i), Qo = Y(/^(?:\w+script|data):/i), $o = Y(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g), es = Y(/^html$/i), ts = Y(/^[a-z][.\w]*(-[.\w]+)+$/i), ns = Y(/<[/\w!]/g), rs = Y(/<[/\w]/g), is = Y(/<\/no(script|embed|frames)/i), as = Y(/\/>/i), $ = {
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
}, os = J(Q({}, [
	"style",
	"script",
	"xmp",
	"iframe",
	"noembed",
	"noframes",
	"plaintext",
	"noscript"
])), ss = function() {
	return typeof window > "u" ? null : window;
}, cs = function(e, t) {
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
}, ls = function() {
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
}, us = function(e, t, n, r) {
	return To(e, t) && uo(e[t]) ? Q(r.base ? Ao(r.base) : {}, e[t], r.transform) : n;
}, ds = function(e, t, n) {
	let r = To(e, t) ? e[t] : void 0;
	return r && typeof r == "object" ? Ao(r) : n();
};
function fs() {
	let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : ss(), t = (e) => fs(e);
	if (t.version = "3.4.17", t.removed = [], !e || !e.document || e.document.nodeType !== $.document || !e.Element) return t.isSupported = !1, t;
	let n = e.document, r = n, i = r.currentScript;
	e.DocumentFragment;
	let a = e.HTMLTemplateElement, o = e.Node, s = e.Element, c = e.NodeFilter;
	e.NamedNodeMap === void 0 && (e.NamedNodeMap || e.MozNamedAttrMap), e.HTMLFormElement;
	let l = e.DOMParser, u = e.trustedTypes, d = s.prototype, f = Mo(d, "cloneNode"), p = Mo(d, "remove"), m = Mo(d, "removeAttributeNode"), h = Mo(d, "nextSibling"), g = Mo(d, "childNodes"), _ = Mo(d, "parentNode"), v = Mo(d, "shadowRoot"), y = Mo(d, "attributes"), b = o && o.prototype ? Mo(o.prototype, "nodeType") : null, x = o && o.prototype ? Mo(o.prototype, "nodeName") : null, S = o && o.prototype ? Mo(o.prototype, "ownerDocument") : null, C = function(e) {
		return b ? b(e) : e.nodeType;
	}, w = function(e) {
		return x ? x(e) : e.nodeName;
	}, ee = No(d, "namespaceURI"), te = function(e) {
		return ee ? ee(e) : e.namespaceURI;
	};
	if (typeof a == "function") {
		let e = n.createElement("template");
		e.content && e.content.ownerDocument && (n = e.content.ownerDocument);
	}
	let T, ne = "", re, ie = !1, ae = 0, oe = function() {
		if (ae > 0) throw Do("A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the \"DOMPurify and Trusted Types\" section of the README.");
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
		return ie ||= (re = cs(u, i), !0), re;
	}, ue = n, de = ue.implementation, fe = ue.createNodeIterator, pe = ue.createDocumentFragment, me = ue.getElementsByTagName, he = r.importNode, E = ls();
	t.isSupported = typeof Ya == "function" && typeof _ == "function" && de && de.createHTMLDocument !== void 0;
	let D = Ko, ge = qo, O = Jo, _e = Yo, ve = Xo, ye = Qo, be = $o, xe = ts, Se = Zo, k = null, Ce = Q({}, [
		...Fo,
		...Io,
		...Lo,
		...zo,
		...Vo
	]), A = null, we = Q({}, [
		...Ho,
		...Uo,
		...Wo,
		...Go
	]), Te = Object.seal(eo(null, {
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
	})), Ee = null, De = null, Oe = Object.seal(eo(null, {
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
	})), ke = !0, Ae = !0, je = !1, Me = !0, Ne = !1, j = !0, Pe = !1, Fe = !1, Ie = null, Le = null, Re = !1, ze = !1, Be = !1, Ve = !1, He = !0, M = !1, Ue = "user-content-", We = !0, Ge = !1, Ke = null, qe = {}, Je = null, Ye = Q({}, /* @__PURE__ */ "annotation-xml.audio.colgroup.desc.foreignobject.head.iframe.math.mi.mn.mo.ms.mtext.noembed.noframes.noscript.plaintext.script.selectedcontent.style.svg.template.thead.title.video.xmp".split(".")), Xe = null, Ze = Q({}, [
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
	]), et = "http://www.w3.org/1998/Math/MathML", tt = "http://www.w3.org/2000/svg", nt = "http://www.w3.org/1999/xhtml", rt = nt, it = !1, at = null, ot = Q({}, [
		et,
		tt,
		nt
	], go), N = J([
		"mi",
		"mo",
		"mn",
		"ms",
		"mtext"
	]), st = Q({}, N), ct = J(["annotation-xml"]), lt = Q({}, ct), ut = Q({}, [
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
		e && typeof e == "object" || (e = {}), e = Ao(e), dt = ft.indexOf(e.PARSER_MEDIA_TYPE) === -1 ? "text/html" : e.PARSER_MEDIA_TYPE, P = dt === "application/xhtml+xml" ? go : ho, k = us(e, "ALLOWED_TAGS", Ce, { transform: P }), A = us(e, "ALLOWED_ATTR", we, { transform: P }), at = us(e, "ALLOWED_NAMESPACES", ot, { transform: go }), Qe = us(e, "ADD_URI_SAFE_ATTR", $e, {
			transform: P,
			base: $e
		}), Xe = us(e, "ADD_DATA_URI_TAGS", Ze, {
			transform: P,
			base: Ze
		}), Je = us(e, "FORBID_CONTENTS", Ye, { transform: P }), Ee = us(e, "FORBID_TAGS", Ao({}), { transform: P }), De = us(e, "FORBID_ATTR", Ao({}), { transform: P }), qe = To(e, "USE_PROFILES") ? e.USE_PROFILES && typeof e.USE_PROFILES == "object" ? Ao(e.USE_PROFILES) : e.USE_PROFILES : !1, ke = e.ALLOW_ARIA_ATTR !== !1, Ae = e.ALLOW_DATA_ATTR !== !1, je = e.ALLOW_UNKNOWN_PROTOCOLS || !1, Me = e.ALLOW_SELF_CLOSE_IN_ATTR !== !1, Ne = e.SAFE_FOR_TEMPLATES || !1, j = e.SAFE_FOR_XML !== !1, Pe = e.WHOLE_DOCUMENT || !1, ze = e.RETURN_DOM || !1, Be = e.RETURN_DOM_FRAGMENT || !1, Ve = e.RETURN_TRUSTED_TYPE || !1, Re = e.FORCE_BODY || !1, He = e.SANITIZE_DOM !== !1, M = e.SANITIZE_NAMED_PROPS || !1, We = e.KEEP_CONTENT !== !1, Ge = e.IN_PLACE || !1, Se = Po(e.ALLOWED_URI_REGEXP) ? e.ALLOWED_URI_REGEXP : Zo, rt = typeof e.NAMESPACE == "string" ? e.NAMESPACE : nt, st = ds(e, "MATHML_TEXT_INTEGRATION_POINTS", () => Q({}, N)), lt = ds(e, "HTML_INTEGRATION_POINTS", () => Q({}, ct));
		let t = ds(e, "CUSTOM_ELEMENT_HANDLING", () => eo(null));
		if (Te = eo(null), To(t, "tagNameCheck") && ht(t.tagNameCheck) && (Te.tagNameCheck = t.tagNameCheck), To(t, "attributeNameCheck") && ht(t.attributeNameCheck) && (Te.attributeNameCheck = t.attributeNameCheck), To(t, "allowCustomizedBuiltInElements") && typeof t.allowCustomizedBuiltInElements == "boolean" && (Te.allowCustomizedBuiltInElements = t.allowCustomizedBuiltInElements), Y(Te), Ne && (Ae = !1), Be && (ze = !0), qe && (k = Q({}, Vo), A = eo(null), qe.html === !0 && (Q(k, Fo), Q(A, Ho)), qe.svg === !0 && (Q(k, Io), Q(A, Uo), Q(A, Go)), qe.svgFilters === !0 && (Q(k, Lo), Q(A, Uo), Q(A, Go)), qe.mathMl === !0 && (Q(k, zo), Q(A, Wo), Q(A, Go))), Oe.tagCheck = null, Oe.attributeCheck = null, To(e, "ADD_TAGS") && (typeof e.ADD_TAGS == "function" ? Oe.tagCheck = e.ADD_TAGS : uo(e.ADD_TAGS) && (k === Ce && (k = Ao(k)), Q(k, e.ADD_TAGS, P))), To(e, "ADD_ATTR") && (typeof e.ADD_ATTR == "function" ? Oe.attributeCheck = e.ADD_ATTR : uo(e.ADD_ATTR) && (A === we && (A = Ao(A)), Q(A, e.ADD_ATTR, P))), To(e, "ADD_FORBID_CONTENTS") && uo(e.ADD_FORBID_CONTENTS) && (Je === Ye && (Je = Ao(Je)), Q(Je, e.ADD_FORBID_CONTENTS, P)), We && (k["#text"] = !0), Pe && Q(k, [
			"html",
			"head",
			"body"
		]), k.table && (Q(k, ["tbody"]), delete Ee.tbody), e.TRUSTED_TYPES_POLICY) {
			if (typeof e.TRUSTED_TYPES_POLICY.createHTML != "function") throw Do("TRUSTED_TYPES_POLICY configuration option must provide a \"createHTML\" hook.");
			if (typeof e.TRUSTED_TYPES_POLICY.createScriptURL != "function") throw Do("TRUSTED_TYPES_POLICY configuration option must provide a \"createScriptURL\" hook.");
			let t = T;
			T = e.TRUSTED_TYPES_POLICY;
			try {
				ne = se("");
			} catch (e) {
				throw T = t, e;
			}
		} else e.TRUSTED_TYPES_POLICY === null ? (T = void 0, ne = "") : (T === void 0 && (T = le()), T && typeof ne == "string" && (ne = se("")));
		J && J(e), pt = e;
	}, _t = Q({}, [
		...Io,
		...Lo,
		...Ro
	]), vt = Q({}, [...zo, ...Bo]), yt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "svg" : t.namespaceURI === et ? e === "svg" && (n === "annotation-xml" || st[n]) : !!_t[e];
	}, bt = function(e, t, n) {
		return t.namespaceURI === nt ? e === "math" : t.namespaceURI === tt ? e === "math" && lt[n] : !!vt[e];
	}, xt = function(e, t, n) {
		return t.namespaceURI === tt && !lt[n] || t.namespaceURI === et && !st[n] ? !1 : !vt[e] && (ut[e] || !_t[e]);
	}, St = function(e) {
		let t = te(e), n = ho(w(e)), r = _(e), i, a;
		return r && C(r) === $.element ? (i = { namespaceURI: te(r) }, a = ho(w(r))) : (i = { namespaceURI: rt }, a = "template"), at[t] ? t === tt ? yt(n, i, a) : t === et ? bt(n, i, a) : t === nt ? xt(n, i, a) : !!(dt === "application/xhtml+xml" && at[t]) : !1;
	}, F = function(e) {
		so(t.removed, { element: e });
		try {
			_(e).removeChild(e);
		} catch {
			if (p(e), !_(e)) throw Do("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
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
		Et(e);
		let t = g(e);
		t && io(co(t), (e) => {
			try {
				p(e);
			} catch {}
		});
		let n = C(e) === $.element ? y(e) : null;
		if (n) for (let t = n.length - 1; t >= 0; --t) {
			let r = n[t], i = r && r.name;
			typeof i == "string" && Ct(e, r, i);
		}
	}, I = function(e, n, r) {
		if (!r) try {
			r = n.getAttributeNode(e);
		} catch {
			r = null;
		}
		so(t.removed, {
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
			if (ze || Be) try {
				F(n);
			} catch {}
			else try {
				n.setAttribute(e, "");
			} catch {}
		}
	}, Tt = function(e) {
		let t = y(e);
		if (t) for (let n = t.length - 1; n >= 0; --n) {
			let r = t[n], i = r && r.name;
			typeof i != "string" || A[P(i)] || Ct(e, r, i);
		}
	}, L = function(e, t) {
		let n = arguments.length > 2 && arguments[2] !== void 0 && arguments[2], r = [e];
		for (; r.length > 0;) {
			let e = r.pop();
			if (t(e) === !1) continue;
			if (n && C(e) === $.element) {
				let t = v(e);
				R(t) && r.push(t);
			}
			let i = g(e);
			if (i) for (let e = i.length - 1; e >= 0; --e) r.push(i[e]);
		}
	}, Et = function(e) {
		L(e, (e) => {
			C(e) === $.element && Tt(e);
		});
	}, Dt = function(e) {
		let t = [];
		return L(e, (e) => {
			t.push(e);
		}, !0), t;
	}, Ot = function(e, t) {
		let n = fo();
		io(Dt(e), (e) => {
			C(e) === $.element && po(n, e);
		}), io(t, (e) => {
			C(e) === $.element && !mo(n, e) && Tt(e);
		});
	}, kt = function(e, n) {
		let r = !1;
		return io(t.removed, (t) => {
			t.element && (r ||= t.element === e, Et(t.element));
		}), Ot(e, n), r;
	}, At = function(e, t) {
		let n = Ke, r = null, i = () => (r || (r = fo(), io(t, (e) => {
			po(r, e);
		})), r);
		for (let t = 0;; t++) {
			let r = Dt(e), a = [e], o = !1;
			if (io(r, (e) => {
				C(e) === $.documentFragment ? so(a, e) : !o && !mo(n, e) && (o = mo(i(), e));
			}), !o) return;
			if (t >= 8) throw Do("the tree kept changing during in-place sanitization; refusing to return it");
			io(a, (t, n) => {
				jt(t, n > 0, e, i());
			});
		}
	}, jt = function(e, t, n, r) {
		let i = It(e), a;
		for (; a = i.nextNode();) mo(r, a) && !mo(Ke, a) && (t && B(E.uponSanitizeShadowNode, a, null), Gt(a, n), Qt(a, n), R(a.content) && $t(a.content));
	}, Mt = function(e) {
		Ke && po(Ke, e);
	}, Nt = function(e, t) {
		return j ? e === "patchsrc" || e === "for" && t !== "label" && t !== "output" : !1;
	}, Pt = function(e) {
		j && L(e, (e) => {
			let t = C(e);
			if (t === $.processingInstruction || t === $.comment && X(rs, e.data)) {
				try {
					p(e);
				} catch {}
				return !1;
			}
			if (t === $.element) {
				let t = e, n = P(w(e));
				try {
					t.hasAttribute && t.hasAttribute("patchsrc") && t.removeAttribute("patchsrc"), t.hasAttribute && t.hasAttribute("for") && Nt("for", n) && t.removeAttribute("for");
				} catch {}
			}
			return !0;
		});
	}, Ft = function(e) {
		let t = null, r = null;
		if (Re) e = "<remove></remove>" + e;
		else {
			let t = _o(e, /^[\r\n\t ]+/);
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
				t.documentElement.innerHTML = it ? ne : i;
			} catch {}
		}
		let a = t.body || t.documentElement;
		return e && r && a.insertBefore(n.createTextNode(r), a.childNodes[0] || null), rt === nt ? me.call(t, Pe ? "html" : "body")[0] : Pe ? t.documentElement : a;
	}, It = function(e) {
		let t = S ? S(e) : e.ownerDocument;
		return fe.call(t || e, e, c.SHOW_ELEMENT | c.SHOW_COMMENT | c.SHOW_TEXT | c.SHOW_PROCESSING_INSTRUCTION | c.SHOW_CDATA_SECTION, null);
	}, Lt = function(e) {
		return e = vo(e, D, " "), e = vo(e, ge, " "), e = vo(e, O, " "), e;
	}, Rt = function(e) {
		zt(e);
		let t = e.querySelectorAll?.call(e, "template");
		t && io(t, (e) => {
			R(e.content) && Rt(e.content);
		});
	}, zt = function(e) {
		e.normalize();
		let t = S ? S(e) : e.ownerDocument, n = fe.call(t || e, e, c.SHOW_TEXT | c.SHOW_COMMENT | c.SHOW_CDATA_SECTION | c.SHOW_PROCESSING_INSTRUCTION, null), r = n.nextNode();
		for (; r;) r.data = Lt(r.data), r = n.nextNode();
	}, Bt = function(e) {
		let t = x ? x(e) : null;
		return typeof t != "string" || P(t) !== "form" ? !1 : typeof e.nodeName != "string" || typeof e.textContent != "string" || typeof e.removeChild != "function" || e.attributes !== y(e) || typeof e.removeAttribute != "function" || typeof e.removeAttributeNode != "function" || typeof e.getAttributeNode != "function" || typeof e.setAttribute != "function" || typeof e.namespaceURI != "string" || typeof e.insertBefore != "function" || typeof e.hasChildNodes != "function" || e.nodeType !== b(e) || e.childNodes !== g(e);
	}, R = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return b(e) === $.documentFragment;
		} catch {
			return !1;
		}
	}, z = function(e) {
		if (!b || typeof e != "object" || !e) return !1;
		try {
			return typeof b(e) == "number";
		} catch {
			return !1;
		}
	};
	function B(e, n, r) {
		e.length !== 0 && io(e, (e) => {
			e.call(t, n, r, pt);
		});
	}
	let Vt = function(e, t) {
		return !!(j && e.hasChildNodes() && !z(e.firstElementChild) && X(ns, e.textContent) && X(ns, e.innerHTML) || j && os[t] && C(e) === $.element && te(e) === nt && (z(e.firstElementChild) || typeof e.textContent == "string" && X(ns, e.textContent)) || e.nodeType === $.processingInstruction || j && e.nodeType === $.comment && X(rs, e.data));
	}, V = function(e, t) {
		return e instanceof RegExp ? X(e, t) : e instanceof Function && !!e(t, ...[...arguments].slice(2));
	}, Ht = function(e, t, n) {
		if (!Ee[t] && Jt(t) && V(Te.tagNameCheck, t)) return !1;
		if (We && !Je[t] && !Je[ho(t)]) {
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
	}, Ut = function(e, t, n, r) {
		return e.length === 0 ? t : t === n || t === r ? Ao(t) : t;
	}, Wt = function(e, t) {
		return e === t || _(e) !== null ? !1 : (Ge && Et(e), !0);
	}, Gt = function(e, n) {
		if (Mt(e), B(E.beforeSanitizeElements, e, null), Wt(e, n)) return !0;
		if (Bt(e)) return F(e), !0;
		let r = P(w(e));
		if (k = Ut(E.uponSanitizeElement, k, Ce, Ie), B(E.uponSanitizeElement, e, {
			tagName: r,
			allowedTags: k
		}), Wt(e, n)) return !0;
		if (Vt(e, r)) return F(e), !0;
		if (Ee[r] || !(Oe.tagCheck instanceof Function && Oe.tagCheck(r)) && !k[r]) {
			let t = Ht(e, r, n);
			return t === !1 && (B(E.afterSanitizeElements, e, null), Wt(e, n)) ? !0 : t;
		}
		if (C(e) === $.element && !St(e) || (r === "noscript" || r === "noembed" || r === "noframes") && X(is, e.innerHTML)) return F(e), !0;
		if (Ne && e.nodeType === $.text) {
			let n = Lt(e.textContent);
			e.textContent !== n && (so(t.removed, { element: e.cloneNode() }), e.textContent = n);
		}
		return B(E.afterSanitizeElements, e, null), Wt(e, n);
	}, Kt = function(e, t, r) {
		if (De[t] || Nt(t, e) || He && (t === "id" || t === "name") && (r in n || r in mt)) return !1;
		let i = A[t] || Oe.attributeCheck instanceof Function && Oe.attributeCheck(t, e);
		return Ae && X(_e, t) || ke && X(ve, t) ? !0 : i ? Qe[t] || X(Se, vo(r, be, "")) || (t === "src" || t === "xlink:href" || t === "href") && e !== "script" && yo(r, "data:") === 0 && Xe[e] || je && !X(ye, vo(r, be, "")) ? !0 : !r : Jt(e) && V(Te.tagNameCheck, e) && V(Te.attributeNameCheck, t, e) || t === "is" && Te.allowCustomizedBuiltInElements && V(Te.tagNameCheck, r);
	}, qt = Q({}, [
		"annotation-xml",
		"color-profile",
		"font-face",
		"font-face-format",
		"font-face-name",
		"font-face-src",
		"font-face-uri",
		"missing-glyph"
	]), Jt = function(e) {
		return !qt[ho(e)] && X(xe, e);
	}, Yt = function(e, t, n, r) {
		if (T && typeof u == "object" && typeof u.getAttributeType == "function" && !n) switch (u.getAttributeType(e, t)) {
			case "TrustedHTML": return se(r);
			case "TrustedScriptURL": return ce(r);
		}
		return r;
	}, Xt = function(e, t, n, r) {
		try {
			return n ? e.setAttributeNS(n, t, r) : e.setAttribute(t, r), !Bt(e) || (F(e), !1);
		} catch {
			return I(t, e), !1;
		}
	}, Zt = function(e, t) {
		return j && X(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i, t) ? !0 : ho(e) === "attributename" && _o(t, "href") !== null;
	}, Qt = function(e, n) {
		if (B(E.beforeSanitizeAttributes, e, null), Wt(e, n)) return;
		let r = e.attributes;
		if (!r || Bt(e)) return;
		A = Ut(E.uponSanitizeAttribute, A, we, Le);
		let i = {
			attrName: "",
			attrValue: "",
			keepAttr: !0,
			allowedAttributes: A,
			forceKeepAttr: void 0
		}, a = r.length, o = P(e.nodeName);
		for (; a--;) {
			let n = r[a], s = n.name, c = n.namespaceURI, l = n.value, u = P(s), d = l, f = s === "value" ? d : bo(d), p = !1;
			i.attrName = u, i.attrValue = f, i.keepAttr = !0, i.forceKeepAttr = void 0, B(E.uponSanitizeAttribute, e, i), f = i.attrValue, M && (u === "id" || u === "name") && yo(f, Ue) !== 0 && (I(s, e, n), f = Ue + f, p = !0), Zt(s, f) ? I(s, e, n) : i.forceKeepAttr || (i.keepAttr ? !Me && X(as, f) ? I(s, e, n) : (Ne && (f = Lt(f)), Kt(o, u, f) ? (f = Yt(o, u, c, f), f !== d && Xt(e, s, c, f) && p && oo(t.removed)) : I(s, e, n)) : I(s, e, n));
		}
		B(E.afterSanitizeAttributes, e, null), Wt(e, n);
	}, $t = function(e) {
		let t = null, n = It(e);
		for (B(E.beforeSanitizeShadowDOM, e, null); t = n.nextNode();) if (B(E.uponSanitizeShadowNode, t, null), Gt(t, e), Qt(t, e), R(t.content) && $t(t.content), C(t) === $.element) {
			let e = v(t);
			R(e) && (en(e), $t(e));
		}
		Ne && zt(e), B(E.afterSanitizeShadowDOM, e, null);
	}, en = function(e) {
		let t = [{
			node: e,
			shadow: null
		}];
		for (; t.length > 0;) {
			let e = t.pop();
			if (e.shadow) {
				$t(e.shadow);
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
					R(e) && t.push({
						node: e,
						shadow: null
					});
				}
			}
			if (r) {
				let e = v(n);
				R(e) && t.push({
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
		if (it = !e, it && (e = "<!-->"), typeof e != "string" && !z(e) && (e = jo(e), typeof e != "string")) throw Do("dirty is not a string, aborting");
		if (!t.isSupported) return e;
		Fe ? (k = Ie, A = Le) : gt(n), (E.uponSanitizeElement.length > 0 || E.uponSanitizeAttribute.length > 0) && (k = Ao(k)), E.uponSanitizeAttribute.length > 0 && (A = Ao(A)), t.removed = [];
		let c = Ge && typeof e != "string" && z(e), l = [], u = Ke;
		if (c) {
			Pt(e);
			let t = w(e);
			if (typeof t == "string") {
				let n = P(t);
				if (!k[n] || Ee[n]) throw wt(e), Do("root node is forbidden and cannot be sanitized in-place");
			}
			if (Bt(e)) throw wt(e), Do("root node is clobbered and cannot be sanitized in-place");
			l = Dt(e), Ke = fo();
		} else if (z(e)) i = Ft("<!---->"), a = i.ownerDocument.importNode(e, !0), a.nodeType === $.element && a.nodeName === "BODY" || a.nodeName === "HTML" ? i = a : i.appendChild(a), en(i);
		else {
			if (!ze && !Ne && !Pe && e.indexOf("<") === -1) return T && Ve ? se(e) : e;
			if (i = Ft(e), !i) return ze ? null : Ve ? ne : "";
		}
		i && Re && F(i.firstChild);
		let d = c ? e : i;
		try {
			c && en(e);
			let t = It(d);
			for (; o = t.nextNode();) Gt(o, d), Qt(o, d), R(o.content) && $t(o.content);
			c && At(e, l);
		} catch (t) {
			throw c && (wt(e), kt(e, l)), t;
		} finally {
			Ke = u;
		}
		if (c) {
			if (kt(e, l)) throw Do("a node selected for removal could not be safely returned; refusing to sanitize in place");
			return Ne && Rt(e), e;
		}
		if (ze) {
			if (Ne && Rt(i), Be) for (s = pe.call(i.ownerDocument); i.firstChild;) s.appendChild(i.firstChild);
			else s = i;
			return (A.shadowroot || A.shadowrootmode) && (s = he.call(r, s, !0)), s;
		}
		let f = Pe ? i.outerHTML : i.innerHTML;
		return Pe && k["!doctype"] && i.ownerDocument && i.ownerDocument.doctype && i.ownerDocument.doctype.name && X(es, i.ownerDocument.doctype.name) && (f = "<!DOCTYPE " + i.ownerDocument.doctype.name + ">\n" + f), Ne && (f = Lt(f)), T && Ve ? se(f) : f;
	}, t.setConfig = function() {
		let e = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
		gt(e), Fe = !0, Ie = k, Le = A;
	}, t.clearConfig = function() {
		pt = null, Fe = !1, Ie = null, Le = null, T = re, ne = "";
	}, t.isValidAttribute = function(e, t, n) {
		pt || gt({});
		let r = P(e), i = P(t);
		return typeof n == "string" && (Zt(t, n) || !Me && X(as, n)) ? !1 : Kt(r, i, n);
	}, t.addHook = function(e, t) {
		typeof t == "function" && To(E, e) && so(E[e], t);
	}, t.removeHook = function(e, t) {
		if (To(E, e)) {
			if (t !== void 0) {
				let n = ao(E[e], t);
				return n === -1 ? void 0 : lo(E[e], n, 1)[0];
			}
			return oo(E[e]);
		}
	}, t.removeHooks = function(e) {
		To(E, e) && (E[e] = []);
	}, t.removeAllHooks = function() {
		E = ls();
	}, t;
}
fs();
// kit/format.js
function ps(e) {
	if (!e || e.startsWith("0001")) return "";
	let t = Math.round((Date.now() - new Date(e).getTime()) / 1e3), n = Math.abs(t), r = t < 0 ? "in " : "", i = t < 0 ? "" : " ago";
	return n < 60 ? r + n + "s" + i : n < 3600 ? r + Math.round(n / 60) + "m" + i : n < 86400 ? r + Math.round(n / 3600) + "h" + i : r + Math.round(n / 86400) + "d" + i;
}
// plugins/sessions/Sessions.svelte
var ms = /* @__PURE__ */ ir("<span class=\"sub\"> </span>"), hs = /* @__PURE__ */ ir("<div class=\"card err\"> </div>"), gs = /* @__PURE__ */ ir("<span class=\"pill bad gap svelte-1yaxaek\"> </span>"), _s = /* @__PURE__ */ ir("<span class=\"pill warn gap svelte-1yaxaek\" title=\"no new claims\"> </span>"), vs = /* @__PURE__ */ ir("<a class=\"mono tk svelte-1yaxaek\"> </a>"), ys = /* @__PURE__ */ ir("<span class=\"pill warn gap svelte-1yaxaek\">override</span>"), bs = /* @__PURE__ */ ir("<button class=\"svelte-1yaxaek\">Uncordon</button>"), xs = /* @__PURE__ */ ir("<button class=\"svelte-1yaxaek\">Cordon</button>"), Ss = /* @__PURE__ */ ir("<tr><td class=\"c-id svelte-1yaxaek\"><a class=\"mono sl\"> </a><!></td><td class=\"c-st svelte-1yaxaek\"><span> </span><!></td><td class=\"svelte-1yaxaek\"><!></td><td class=\"muted svelte-1yaxaek\"> </td><td class=\"muted mono svelte-1yaxaek\"> <!></td><td class=\"muted c-time svelte-1yaxaek\"> </td><td class=\"c-act svelte-1yaxaek\"><!></td></tr>"), Cs = /* @__PURE__ */ ir("<tr><td colspan=\"7\" class=\"faint\">No live sessions.</td></tr>"), ws = /* @__PURE__ */ ir("<!> <div class=\"card scroll\"><table class=\"stack\" data-testid=\"sessions\"><thead><tr><th>Session</th><th>State</th><th>Ticket</th><th>Pool</th><th>Model</th><th>Last event</th><th></th></tr></thead><tbody></tbody></table></div>", 1), Ts = /* @__PURE__ */ ir("<header class=\"vh\"><h1>Sessions</h1> <!></header> <div class=\"pad\"><!></div>", 1);
function Es(e, t) {
	Fe(t, !0);
	let n = Hr(() => t.api), r = (e) => t.api.go(e), i = () => t.api.closeSession(), a = Vr(() => t.api, () => "api/sessions"), o = /* @__PURE__ */ I("");
	async function s(e, n) {
		e.stopPropagation(), L(o, "");
		try {
			await t.api.post("api/session/cordon", n), a.refresh();
		} catch (e) {
			L(o, Ur(e), !0);
		}
	}
	function c(e) {
		return e.alive ? e.backoff_until ? ["backoff", "warn"] : e.drained ? ["drained", "warn"] : e.current ? ["working", "ok"] : ["idle", ""] : ["dead", "bad"];
	}
	var l = Ts(), u = B(l), d = V(z(u), 2), f = (e) => {
		var t = ms(), n = Vt(t);
		an(() => pr(n, `${a.data.length ?? ""} live`)), or(e, t);
	};
	br(d, (e) => {
		a.data && e(f);
	}), ye(u);
	var p = V(u, 2), m = z(p), h = (e) => {
		var t = hs(), n = Vt(t, !0);
		an(() => pr(n, a.error)), or(e, t);
	}, g = (e) => {
		var t = ws(), l = B(t), u = (e) => {
			var t = hs(), n = Vt(t, !0);
			an(() => pr(n, W(o))), or(e, t);
		};
		br(l, (e) => {
			W(o) && e(u);
		});
		var d = V(l, 2), f = z(d), p = V(z(f));
		wr(p, 21, () => a.data, (e) => e.id, (e, t) => {
			let a = /* @__PURE__ */ $e(() => c(W(t)));
			var o = Ss();
			let l;
			var u = z(o), d = z(u), f = Vt(d, !0), p = V(d), m = (e) => {
				var n = gs(), r = Vt(n);
				an(() => pr(r, `${W(t).crashes ?? ""} crashes`)), or(e, n);
			};
			br(p, (e) => {
				W(t).crashes && e(m);
			}), ye(u);
			var h = V(u), g = z(h), _ = Vt(g, !0), v = V(g), y = (e) => {
				var n = _s(), r = Vt(n);
				an(() => pr(r, `cordoned: ${W(t).cordon ?? ""}`)), or(e, n);
			};
			br(v, (e) => {
				W(t).cordon && e(y);
			}), ye(h);
			var b = V(h), x = z(b), S = (e) => {
				var n = vs(), i = Vt(n, !0);
				an(() => {
					Ir(n, "href", "?ticket=" + W(t).current), pr(i, W(t).current);
				}), Yn("click", n, (e) => {
					e.preventDefault(), e.stopPropagation(), r({ ticket: W(t).current });
				}), or(e, n);
			};
			br(x, (e) => {
				W(t).current && e(S);
			}), ye(b);
			var C = V(b), w = Vt(C, !0), ee = V(C), te = z(ee), T = V(te), ne = (e) => {
				or(e, ys());
			};
			br(T, (e) => {
				W(t).override && e(ne);
			}), ye(ee);
			var re = V(ee), ie = Vt(re, !0), ae = V(re), oe = z(ae), se = (e) => {
				var n = ar(), r = B(n), i = (e) => {
					var n = bs();
					Yn("click", n, (e) => s(e, {
						id: W(t).id,
						uncordon: !0
					})), or(e, n);
				}, a = (e) => {
					var n = xs();
					Yn("click", n, (e) => s(e, { id: W(t).id })), or(e, n);
				};
				br(r, (e) => {
					W(t).cordon ? e(i) : e(a, -1);
				}), or(e, n);
			};
			br(oe, (e) => {
				W(t).alive && e(se);
			}), ye(ae), ye(o), an((e) => {
				l = Mr(o, 1, "click svelte-1yaxaek", null, l, { sel: n.session === W(t).id }), Ir(d, "href", "?session=" + W(t).id), pr(f, W(t).id), Mr(g, 1, `pill ${W(a)[1] ?? ""}`, "svelte-1yaxaek"), pr(_, W(a)[0]), pr(w, W(t).ticket?.pool ?? ""), pr(te, `${W(t).model ?? "" ?? ""}${W(t).effort ? " · " + W(t).effort : ""}`), pr(ie, e);
			}, [() => ps(W(t).seq_time)]), Yn("click", o, () => n.session === W(t).id ? i() : r({ session: W(t).id })), Yn("click", d, (e) => {
				e.preventDefault(), e.stopPropagation(), r({
					session: W(t).id,
					ticket: ""
				});
			}), or(e, o);
		}, (e) => {
			or(e, Cs());
		}), ye(p), ye(f), ye(d), or(e, t);
	};
	br(m, (e) => {
		a.error ? e(h) : a.data && e(g, 1);
	}), ye(p), or(e, l), Ie();
}
// plugins/sessions/index.js
Xn(["click"]), Br("desk-sessions-view", Es);
