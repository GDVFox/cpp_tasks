#include <iostream>
#include <cmath>
#include "Function.h"

class TestClassF {
public:
    TestClassF() {
        std::cout << "TestClassF created!" << std::endl;
    };
    int operator()(double b) { return (int)floor(b); }
    ~TestClassF() {
        std::cout << "TestClassF deleted!" << std::endl;
    }
};

class TestClassG {
public:
    TestClassG() {
        std::cout << "TestClassG created!" << std::endl;
    };
    std::string operator()(int b) { return std::to_string(b) + ":G converted!"; }
    ~TestClassG() {
        std::cout << "TestClassG deleted!" << std::endl;
    }
};

class TestClassM {
public:
    explicit TestClassM(std::string add) : add(add) {
        std::cout << "TestClassM created!" << std::endl;
    };
    std::string operator()(std::string b) { return b + ":M!:" + add; }
    ~TestClassM() {
        std::cout << "TestClassM deleted!" << std::endl;
    }

private:
    std::string add;
};


int main() {
    TestClassF testF;
    TestClassG testG;
    TestClassM testM("666");
    TestClassM testM1("13");
    Function<double, int, TestClassF>f(testF);
    Function<int, std::string, TestClassG>g(testG);
    Function<std::string, std::string, TestClassM>m(testM);
    Function<std::string, std::string, TestClassM>m1(testM1);

    std::cout << f(5.5) << std::endl;
    std::cout << g(666) << std::endl;
    std::cout << m("11") << std::endl;
    std::cout << m1("11") << std::endl;

    std::cout << (m*(g*f))(6.5) << std::endl;
    std::cout << ((m1*m)*g)(777) << std::endl;

    std::cout << (m1*m*g)(700) << std::endl;
    std::cout << ((m1*m)*(g*f))(666.13) << std::endl;

    return 0;
}