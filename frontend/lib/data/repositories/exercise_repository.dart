import '../../data/api/api_client.dart';
import '../../data/api/exercise_api.dart';
import '../../data/mock/mock_data.dart';
import '../../data/models/models.dart';

/// Repository for exercise catalog data.
///
/// Fetches from the API when authenticated and reachable; falls back to
/// mock data when offline or unauthenticated. UI code calls this instead
/// of reading [MockData] directly.
class ExerciseRepository {
  final ExerciseApi _api;

  ExerciseRepository(ApiClient client) : _api = ExerciseApi(client);

  /// Returns the full exercise catalog.
  ///
  /// Tries the API first. Falls back to mock data on any error (no network,
  /// unauthenticated, server error).
  Future<List<Exercise>> fetchAll() async {
    try {
      final result = await _api.list(limit: 200);
      return result.items;
    } catch (e) {
      // Fall back to mock data during development or when offline
      return List.unmodifiable(MockData.exerciseCatalog);
    }
  }

  /// Search exercises by name/muscle group/equipment.
  Future<List<Exercise>> search({
    String? query,
    String? muscleGroup,
    String? equipment,
    int limit = 20,
  }) async {
    try {
      final result = await _api.list(
        search: query,
        muscleGroup: muscleGroup,
        equipment: equipment,
        limit: limit,
      );
      return result.items;
    } catch (_) {
      // Local filter on mock data as fallback
      var results = MockData.exerciseCatalog;
      if (query != null && query.isNotEmpty) {
        final q = query.toLowerCase();
        results = results.where((e) => e.name.toLowerCase().contains(q)).toList();
      }
      if (muscleGroup != null && muscleGroup.isNotEmpty) {
        results = results.where((e) =>
            e.muscleGroup.toLowerCase() == muscleGroup.toLowerCase()).toList();
      }
      if (equipment != null && equipment.isNotEmpty) {
        results = results.where((e) =>
            e.equipment.toLowerCase() == equipment.toLowerCase()).toList();
      }
      return results;
    }
  }

  /// Get a single exercise by ID. Falls back to mock lookup.
  Future<Exercise?> getById(String id) async {
    try {
      return await _api.getById(id);
    } catch (_) {
      try {
        return MockData.exerciseCatalog.firstWhere((e) => e.id == id);
      } catch (_) {
        return null;
      }
    }
  }
}
