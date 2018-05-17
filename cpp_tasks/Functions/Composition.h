//
// Created by gdvfox on 11.05.18.
//

#ifndef WORK_9_COMPOSITION_H
#define WORK_9_COMPOSITION_H

#include "AbstractFunction.h"

template <class A, class B, class R>
class Composition : public AbstractFunction<A, R> {
    template <class A_1, class B_1, class R_1>
    friend Composition<B_1, A_1, R_1> operator*(const AbstractFunction<A_1, R_1> &f, const AbstractFunction<B_1, A_1> &g);
public:
    Composition(const AbstractFunction<A, B> &f, const AbstractFunction<B, R> &g) : f(f), g(g) {};
    virtual R operator()(A arg) const;
private:
    const AbstractFunction<A, B> &f;
    const AbstractFunction<B, R> &g;
};

template<class A, class B, class R>
R Composition<A, B, R>::operator()(A arg) const {
    return g(f(arg));
}

template <class A_1, class B_1, class R_1>
Composition<B_1, A_1, R_1> operator*(const AbstractFunction<A_1, R_1> &f, const AbstractFunction<B_1, A_1> &g) {
    return Composition<B_1, A_1, R_1>(g, f);
}


#endif //WORK_9_COMPOSITION_H
