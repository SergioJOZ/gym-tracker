import '../../data/api/api_client.dart';
import '../../data/models/models.dart';

/// API client for exercise catalog endpoints.
class ExerciseApi {
  final ApiClient _client;

  ExerciseApi(this._client);

  /// GET /exercises?search=&muscle_group=&equipment=&cursor=&limit=
  Future<({List<Exercise> items, String? nextCursor, bool hasMore})> list({
    String? search,
    String? muscleGroup,
    String? equipment,
    String? difficulty,
    String? category,
    String? cursor,
    int limit = 20,
  }) async {
    final params = <String, dynamic>{
      if (search != null && search.isNotEmpty) 'search': search,
      if (muscleGroup != null && muscleGroup.isNotEmpty) 'muscle_group': muscleGroup,
      if (equipment != null && equipment.isNotEmpty) 'equipment': equipment,
      if (difficulty != null && difficulty.isNotEmpty) 'difficulty': difficulty,
      if (category != null && category.isNotEmpty) 'category': category,
      if (cursor != null && cursor.isNotEmpty) 'cursor': cursor,
      'limit': limit,
    };

    final response = await _client.dio.get('/exercises', queryParameters: params);

    final data = response.data as Map;
    final items = (data['items'] as List)
        .map((e) => Exercise.fromJson(e as Map<String, dynamic>))
        .toList();

    return (
      items: items,
      nextCursor: data['next_cursor'] as String?,
      hasMore: data['has_more'] as bool? ?? false,
    );
  }

  /// GET /exercises/{id}
  Future<Exercise> getById(String id) async {
    final response = await _client.dio.get('/exercises/$id');
    return Exercise.fromJson(response.data as Map<String, dynamic>);
  }
}
