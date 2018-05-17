#include <iostream>
#include <vector>
#include <algorithm>

using namespace std;

bool dfs(int s, int state, vector<vector<int>> &incomp,
        vector<int> &color, pair<vector<int>,vector<int>> &part) {
    color[s] = state;

    if (state == 1) part.first.push_back(s);
    else part.second.push_back(s);

    for (int i = 0; i < incomp[s].size(); i++) {
        if (color[incomp[s][i]] == 0) {
            if (state == 1) {
                if (!dfs(incomp[s][i], 2, incomp, color, part)) return false;
            }
            else {
                if (!dfs(incomp[s][i], 1, incomp, color, part)) return false;
            }
        } else if (color[incomp[s][i]] == state) {
            return false;
        }
    }

    return true;
}

vector<int> concatVectors(vector<int> &src1, vector<int> &src2) {
    vector<int> dest;
    dest.reserve(src1.size() + src2.size());
    dest.insert(dest.end(), src1.begin(), src1.end());
    dest.insert(dest.end(), src2.begin(), src2.end());

    return dest;
}

pair<vector<int>, vector<int>> findOptimal(int c,
               vector<pair<vector<int>,vector<int>>> &parts, pair<vector<int>, vector<int>> result) {

    if (c == parts.size()) return result;

    // Сначала не поворачиваем компоненту
    pair<vector<int>, vector<int>>firstVariant = findOptimal(c + 1, parts,
                                                   make_pair(concatVectors(result.first, parts[c].first),
                                                            concatVectors(result.second, parts[c].second)));

    // Поворачиваем компоненту
    pair<vector<int>, vector<int>>secondVariant = findOptimal(c + 1, parts,
                                                   make_pair(concatVectors(result.first, parts[c].second),
                                                            concatVectors(result.second, parts[c].first)));

    // Смотрим разницу разностей разбиения.
    int delta = abs((int)firstVariant.first.size() - (int)firstVariant.second.size()) -
                abs((int)secondVariant.first.size() - (int)secondVariant.second.size());

    if (delta < 0) {
        return firstVariant;
    } else if (delta == 0) {
        // Выбираем меньшую группу в первом варианте
        // и подбираем соответсвующую группу для сравнения из второго варианта

        if (firstVariant.first.size() < firstVariant.second.size()) {
            if ((secondVariant.first.size() < secondVariant.second.size()
                 && firstVariant.first <= secondVariant.first)
                    || (secondVariant.first.size() >= secondVariant.second.size()
                        && firstVariant.first <= secondVariant.second)) {
                return firstVariant;
            } else {
                return secondVariant;
            }
        } else {
            if ((secondVariant.first.size() < secondVariant.second.size()
                 && firstVariant.second <= secondVariant.first)
                    || (secondVariant.first.size() >= secondVariant.second.size()
                        && firstVariant.second <= secondVariant.second)) {
                return firstVariant;
            } else {
                return secondVariant;
            }
        }
    } else {
        return secondVariant;
    }
}

int main(int argc, char **argv) {
    int n;
    cin >> n;

    vector<vector<int>>incomp(n);               // Пердставление графа несовметимости
    char ch;
    for (int i = 0; i < n; i++) {
        for (int j = 0; j < n; j++) {
            cin >> ch;
            if (ch == '+') {
                incomp[i].push_back(j);
            }
        }
    }


    // Разбиваем граф на компоненты связности,
    // а компоненты связности на доли.
    vector<int>color(n, 0);                     // Вспомогательный массив
                                                // для разбиения на доли.

    vector<pair<vector<int>,vector<int>>>parts; // Хранит разбиения компонент на доли.

    for (int i = 0; i < n; i++) {
        if (color[i] == 0) {
            pair<vector<int>, vector<int>>part;
            if (!dfs(i, 1, incomp, color, part)) {
                cout << "No solution";
                return 0;
            };

            parts.push_back(part);
        }
    }

    pair<vector<int>, vector<int>> optimal = findOptimal(0, parts,
                                                         make_pair(vector<int>(), vector<int>()));

    sort(optimal.first.begin(), optimal.first.end());
    sort(optimal.second.begin(), optimal.second.end());

    if (optimal.first.size() < optimal.second.size()
        || (optimal.first.size() == optimal.second.size() && optimal.first < optimal.second)) {

        for (int i = 0; i < optimal.first.size(); i++) {
            cout << optimal.first[i] + 1 << ' ';
        }

    } else if (optimal.first.size() > optimal.second.size()
            || (optimal.first.size() == optimal.second.size() && optimal.first > optimal.second)) {

        for (int i = 0; i < optimal.second.size(); i++) {
            cout << optimal.second[i] + 1 << ' ';
        }

    }

    return 0;
}