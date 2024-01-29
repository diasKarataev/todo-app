import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './App.css';

const App = () => {
  const [tasks, setTasks] = useState([]);
  const [newTask, setNewTask] = useState({ name: '', details: '' });
  const [editingTask, setEditingTask] = useState(null);
  const [pagination, setPagination] = useState({ page: 1, pageSize: 5 });
  const [filters, setFilters] = useState({
    name: '',
    details: '',
    star: false,
    sortField: '',
    sortOrder: 'asc',
  });

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await axios.get('http://localhost:8000/tasks', {
          params: {
            page: pagination.page,
            pageSize: pagination.pageSize,
            ...filters,
            sortOrder: filters.sortOrder, // Include sortOrder directly
          },
        });

        setTasks(response.data);
        setPagination((prev) => ({ ...prev, total: response.headers['x-total-count'] }));
      } catch (error) {
        console.error('Ошибка получения задач:', error);
      }
    };

    // Check if pagination has changed before making a request
    if (
        pagination.page !== 1 ||
        pagination.pageSize !== 5 ||
        tasks.length === 0 // Add this condition to avoid fetching tasks when tasks are already present
    ) {
      fetchTasks();
    }
  }, [pagination, tasks]);  // Include tasks in the dependencies array


  const fetchTasks = async () => {
    try {
      const response = await axios.get('http://localhost:8000/tasks', {
        params: {
          page: pagination.page,
          pageSize: pagination.pageSize,
          ...filters,
          sortOrder: filters.sortOrder, // Include sortOrder directly
        },
      });

      setTasks(response.data);
      setPagination((prev) => ({ ...prev, total: response.headers['x-total-count'] }));
    } catch (error) {
      console.error('Ошибка получения задач:', error);
    }
  };

  const handleFilterChange = (event) => {
    const { name, value, type } = event.target;

    // Handle different input types (text, checkbox, select, etc.)
    const newValue = type === 'checkbox' ? !filters[name] : value;

    setFilters((prevFilters) => ({
      ...prevFilters,
      [name]: newValue,
    }));
  };

  const handleSortChange = (sortField) => {
    setFilters((prevFilters) => ({
      ...prevFilters,
      sortField,
    }));
  };

  const handleSortOrderChange = (sortOrder) => {
    setFilters((prevFilters) => ({
      ...prevFilters,
      sortOrder,
    }));
  };

  const handleInputChange = (event) => {
    const { name, value } = event.target;
    setNewTask({ ...newTask, [name]: value });
  };
  const handleEdit = (task) => {
    setEditingTask(task); // Устанавливаем задачу для редактирования
  };

  const handleUpdate = async () => {
    try {
      const response = await axios.put(`http://localhost:8000/tasks/${editingTask.ID}`, editingTask);
      const updatedTasks = tasks.map((task) => (task.ID === editingTask.ID ? response.data : task));
      setTasks(updatedTasks);
      setEditingTask(null); // Сбрасываем редактирование
    } catch (error) {
      console.error('Ошибка обновления задачи:', error);
    }
  };
  const handleSubmit = async (event) => {
    event.preventDefault();

    try {
      const response = await axios.post('http://localhost:8000/tasks', newTask);
      setTasks([...tasks, response.data]);
      setNewTask({ name: '', details: '' });
    } catch (error) {
      console.error('Ошибка создания задачи:', error);
    }
  };

  const handleEditedInputChange = (event) => {
    const { name, value } = event.target;
    setEditingTask({ ...editingTask, [name]: value });
  };

  const handleCancelEdit = () => {
    setEditingTask(null); // Отменяем редактирование
  };

  const handleToggleStar = async (taskId) => {
    try {
      const response = await axios.patch(`http://localhost:8000/tasks/${taskId}/toggle-star`);
      const { haveStar, lastUpdated } = response.data;

      // Update tasks state to reflect the star change
      setTasks((prevTasks) =>
          prevTasks.map((task) =>
              task.ID === taskId ? { ...task, star: haveStar, lastUpdated } : task
          )
      );
    } catch (error) {
      console.error('Ошибка изменения статуса звезды:', error);
    }
  };


  const handleFilterApply = () => {
    // Trigger a fetch with the current filters
    fetchTasks();
  };

  const handleDelete = async (taskId) => {
    try {
      await axios.delete(`http://localhost:8000/tasks/${taskId}`);
      const updatedTasks = tasks.filter((task) => task.ID !== taskId);
      setTasks(updatedTasks);
    } catch (error) {
      console.error('Ошибка удаления задачи:', error);
    }
  };

  const handlePageChange = (newPage) => {
    setPagination(prev => ({ ...prev, page: newPage }));
  };

  return (
      <div className='container'>
        <h1>Добавить задачу</h1>
        <form onSubmit={handleSubmit}>
          <input
              type="text"
              name="name"
              placeholder="Название задачи"
              value={newTask.name}
              onChange={handleInputChange}
          />
          <input
              type="text"
              name="details"
              placeholder="Детали"
              value={newTask.details}
              onChange={handleInputChange}
          />
          <button type="submit">Добавить задачу</button>
        </form>

        <h1>Список задач</h1>

        {/* Filter controls */}
        <div>
          <label>
            Name:
            <input type='text' name='name' value={filters.name} onChange={handleFilterChange}/>
          </label>
          <label>
            Details:
            <input type='text' name='details' value={filters.details} onChange={handleFilterChange}/>
          </label>
          <label>
            Star:
            <input type='checkbox' name='star' checked={filters.star} onChange={handleFilterChange}/>
          </label>
          <p></p>
          <button onClick={handleFilterApply}>Применить фильтр</button>

          <p></p>
          {/* Sorting controls */}
          <label>
            Sort by:
            <select name='sortField' value={filters.sortField} onChange={(e) => handleSortChange(e.target.value)}>
              <option value=''>None</option>
              <option value='name'>Name</option>
              <option value='details'>Details</option>
              <option value='lastUpdated'>Updated time</option>
              {/* Add more options based on your task structure */}
            </select>

            <select name='sortOrder' value={filters.sortOrder} onChange={(e) => handleSortOrderChange(e.target.value)}>
              <option value='asc'>По возрастанию</option>
              <option value='desc'>По убыванию</option>
              {/* Add more options based on your task structure */}
            </select>

            <p></p>

            <button onClick={handleFilterApply}>Отсортировать</button>
          </label>

          <ul>
            {tasks.map((task) => (
                <li key={task.ID}>
                  {editingTask && editingTask.ID === task.ID ? (
                      <div>
                        <input type="text" name="name" value={editingTask.name} onChange={handleEditedInputChange}/>
                        <input type="text" name="details" value={editingTask.details}
                               onChange={handleEditedInputChange}/>
                        <button onClick={handleUpdate}>Сохранить</button>
                        <button className='cancel-btn' onClick={handleCancelEdit}>Отмена</button>
                      </div>
                  ) : (
                      <div>
                        <h2>{task.name}</h2>
                        <p>ID: {task.ID}</p>
                        <p>{task.details}</p>
                        <p>Created: {task.createdDate}</p>
                        <p>Last updated: {task.lastUpdated}</p>
                        <button onClick={() => handleEdit(task)}>Редактировать</button>
                        <button className='delete-btn' onClick={() => handleDelete(task.ID)}>Удалить</button>
                        <button className='star-btn' onClick={() => handleToggleStar(task.ID)}>
                          {task.star ? 'Убрать звезду' : 'Поставить звезду'}
                        </button>
                        {task.star ? (
                            <span role="img" aria-label="filled-star">⭐️</span>
                        ) : (
                            <span role="img" aria-label="empty-star">☆</span>
                        )}
                      </div>
                  )}
                </li>
            ))}
          </ul>

          {/* Pagination controls */}
          <div>
            <span>Page {pagination.page}</span>
            <button onClick={() => handlePageChange(pagination.page - 1)} disabled={pagination.page === 1}>
              Previous Page
            </button>
            <button
                onClick={() => handlePageChange(pagination.page + 1)}
                disabled={pagination.page === Math.ceil(pagination.total / pagination.pageSize)}
            >
              Next Page
            </button>
          </div>
        </div>
      </div>
  );
};

export default App;
