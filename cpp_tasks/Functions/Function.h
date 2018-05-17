//
// Created by gdvfox on 11.05.18.
//

#ifndef WORK_9_FUNCTION_H
#define WORK_9_FUNCTION_H

#include "Composition.h"

template<class A, class R, class F>
class Function : public AbstractFunction<A, R> {
    template <class A_1, class B_1, class R_1>
    friend Composition<B_1, A_1, R_1> operator*(const AbstractFunction<A_1, R_1> &f, const AbstractFunction<B_1, A_1> &g);
public:
    explicit Function(F &func) : func(func) {};
    virtual R operator()(A arg) const;
private:
    F &func;
};

template<class A, class R, class F>
R Function<A, R, F>::operator()(A arg) const {
    return func(arg);
}

#endif //WORK_9_FUNCTION_H
