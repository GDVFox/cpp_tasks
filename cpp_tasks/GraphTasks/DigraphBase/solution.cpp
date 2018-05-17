#include <iostream>
#include <vector>
#include <stack>

using namespace std;

/** Структура представляющая вершину орграфа
 *  и атрибутами для алгоритма Тарьяна */
struct Vertex {
    int comp;   // Номер компоненты, которой принадлежит

    int low;    // Минимальное время захода для всех узов,
                // достижимых из вершины и не принадл. выч. компонентам

    int timeIn; // Время захода
};

int graphTime = 1;  // Время для обхода(кол-во шагов)
int compCount = 1;  // Счетчик компонент

/** Поиск в ширину про графу, так же находит минимальные компоненты */
void dfs(int s, vector<vector<int>> &graph, vector<Vertex> &vertexes,
         vector<bool> &used, vector<bool> &inBase) {
    used[s] = true;

    for (int u : graph[s]) {
        // Вершина конденсации достжима из другой вершины =>
        // => не принадлежит базе.
        if (vertexes[s].comp != vertexes[u].comp) {
            inBase[vertexes[u].comp] = false;
        }

        if (!used[u]) {
            dfs(u, graph, vertexes, used, inBase);
        }
    }
}

/** Алгоритм Тарьяна для поиска компонент
 *  сильной связности в орграфе. */
void dfsTarjan(int s, vector<vector<int>> &graph,
               vector<Vertex> &vertexes, stack<int> &vs) {
    vertexes[s].timeIn = graphTime;
    vertexes[s].low = graphTime;
    graphTime++;
    vs.push(s);

    for (int u : graph[s]) {
        // Вершина ещё не была посещена
        if (vertexes[u].timeIn == 0) {
            dfsTarjan(u, graph, vertexes, vs);
        }

        // Обновляем минимальное время захода для достижимых вершин.
        if (vertexes[u].comp == 0
            && vertexes[s].low > vertexes[u].low) {
            vertexes[s].low = vertexes[u].low;
        }
    }

    // Нашли корень новой компоненты сильной связности,
    // добавили все, в которых были пройдены после неё.
    if (vertexes[s].timeIn == vertexes[s].low) {
        int u;
        do {
            u = vs.top(); vs.pop();
            vertexes[u].comp = compCount;
        } while (u != s);

        compCount++;
    }
}

/** Поиск базы в орграфе, возвращает вектор
 * с номерами вершин базы в порядке возрастания.
 * При этом номера выбранных вершин базы минимальны. */
vector<int> findBase(vector<vector<int>> &graph, vector<Vertex> &vertexes) {
    //Поиск компонент сильной связности алгоритмом Тарьяна.
    stack<int>vertexStack;
    for (int i = 0; i < graph.size(); i++) {
        if (!vertexes[i].timeIn) {
            dfsTarjan(i, graph, vertexes, vertexStack);
        }
    }

    // Поиск базы в конденсации орграфа
    vector<bool>used(graph.size(), false);
    vector<bool>inBase(compCount, true);
    for (int i = 0; i < vertexes.size(); i++) {
        if (!used[i]) {
            dfs(i, graph, vertexes, used, inBase);
        }
    }

    // Получение вектора базовых вершин
    // с мин. номерами в порядке возрастания.
    vector<int>base;
    for (int i = 0; i < vertexes.size(); i++) {
        if (inBase[vertexes[i].comp]) {
            base.push_back(i);
            inBase[vertexes[i].comp] = false;
        }
    }

    return base;
}

int main(int argc, char **argv) {
    int n, m;
    cin >> n >> m;

    vector<vector<int>>graph(n);
    vector<Vertex>vertexes(n);

    int x, y;
    for (int i = 0; i < m; i++) {
        cin >> x >> y;
        graph[x].push_back(y);
    }

    for (int v : findBase(graph, vertexes)) {
        cout << v << ' ';
    }

    return 0;
}
