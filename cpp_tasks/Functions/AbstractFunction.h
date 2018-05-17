//
// Created by gdvfox on 11.05.18.
//

#ifndef WORK_9_ABSTRACTFUNCTION_H
#define WORK_9_ABSTRACTFUNCTION_H


template <class A, class R>
class AbstractFunction {
public:
    virtual R operator()(A arg) const = 0;
};

#endif //WORK_9_ABSTRACTFUNCTION_H
